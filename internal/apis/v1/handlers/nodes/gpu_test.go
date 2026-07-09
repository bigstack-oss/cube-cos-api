package nodes

import (
	"errors"
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	nvmlmock "github.com/NVIDIA/go-nvml/pkg/nvml/mock"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/stretchr/testify/require"
)

const bytesPerMiB = uint64(1024 * 1024)

// Restores the test seams after each test so stubs do not leak between tests.
func restoreGpuSeams(t *testing.T) {
	t.Helper()

	origGetNodeGpusMap := getNodeGpusMap
	origGetNodeVgpuProfilesMap := getNodeVgpuProfilesMap
	origGetNodePgpuAttachedInstance := getNodePgpuAttachedInstance
	origDeviceGetHandleByUUID := deviceGetHandleByUUID
	origBuildInstanceLinks := buildInstanceLinks
	origGetOpenstackServer := getOpenstackServer
	origCreateConsole := createConsole
	origVgpuInstanceId := vgpuInstanceId

	t.Cleanup(func() {
		getNodeGpusMap = origGetNodeGpusMap
		getNodeVgpuProfilesMap = origGetNodeVgpuProfilesMap
		getNodePgpuAttachedInstance = origGetNodePgpuAttachedInstance
		deviceGetHandleByUUID = origDeviceGetHandleByUUID
		buildInstanceLinks = origBuildInstanceLinks
		getOpenstackServer = origGetOpenstackServer
		createConsole = origCreateConsole
		vgpuInstanceId = origVgpuInstanceId
	})
}

func TestListLocalGpuCardsHexError(t *testing.T) {
	restoreGpuSeams(t)

	getNodeGpusMap = func(nodeName string) (map[string]gpu.GpuFromHex, error) {
		return nil, errors.New("hex_sdk failed")
	}

	h := &helper{node: "node-1"}
	cards, err := h.listLocalGpuCards()

	require.Error(t, err)
	require.Nil(t, cards)
}

func TestListLocalGpuCardsIncludesGpusInvisibleToNvml(t *testing.T) {
	restoreGpuSeams(t)

	visibleUUID := "GPU-11111111-1111-1111-1111-111111111111"
	passthroughUUID := "GPU-22222222-2222-2222-2222-222222222222"
	reservedUUID := "GPU-33333333-3333-3333-3333-333333333333"

	getNodeGpusMap = func(nodeName string) (map[string]gpu.GpuFromHex, error) {
		return map[string]gpu.GpuFromHex{
			"0000:02:00.0": {
				Id:         visibleUUID,
				Name:       "NVIDIA A100",
				Type:       gpu.ResourceTypePgpu,
				PciAddress: "0000:02:00.0",
				Status:     gpu.GpuStatusIdle,
			},
			"0000:01:00.0": {
				Id:         passthroughUUID,
				Name:       "NVIDIA A100",
				Type:       gpu.ResourceTypePgpu,
				PciAddress: "0000:01:00.0",
				Status:     gpu.GpuStatusInUse,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
			// Reserved for passthrough (vfio-bound) but not attached to any VM
			// yet: invisible to NVML while having no allocation.
			"0000:03:00.0": {
				Id:         reservedUUID,
				Name:       "NVIDIA A100",
				Type:       gpu.ResourceTypePgpu,
				PciAddress: "0000:03:00.0",
				Status:     gpu.GpuStatusIdle,
			},
		}, nil
	}

	getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
		return &gpu.PgpuAttachedInstanceFromHex{Id: "vm-1", Name: "instance-1"}, nil
	}

	buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
		return gpu.InstanceLinks{}
	}

	deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
		if uuid != visibleUUID {
			// A vfio-bound GPU is invisible to NVML.
			return nil, nvml.ERROR_NOT_FOUND
		}

		return &nvmlmock.Device{
			GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
				return nvml.Memory{
					Used:  2048 * bytesPerMiB,
					Total: 8192 * bytesPerMiB,
				}, nvml.SUCCESS
			},
			GetUtilizationRatesFunc: func() (nvml.Utilization, nvml.Return) {
				return nvml.Utilization{Gpu: 55, Memory: 40}, nvml.SUCCESS
			},
		}, nvml.SUCCESS
	}

	h := &helper{node: "node-1"}
	cards, err := h.listLocalGpuCards()

	require.NoError(t, err)
	require.Len(t, cards, 3)

	// Cards are sorted by PCI address.
	passthroughCard, visibleCard, reservedCard := cards[0], cards[1], cards[2]

	require.Equal(t, passthroughUUID, passthroughCard.Id)
	require.Equal(t, "0000:01:00.0", passthroughCard.PciAddress)
	require.Equal(t, gpu.GpuStatusInUse, passthroughCard.Status.Current)
	// NVML data is unavailable for a passthrough GPU; hex data is still reported.
	require.Equal(t, gpu.VramInfo{}, passthroughCard.Vram)
	require.Equal(t, gpu.GpuInfo{}, passthroughCard.Gpu)

	require.Equal(t, visibleUUID, visibleCard.Id)
	require.Equal(t, "0000:02:00.0", visibleCard.PciAddress)
	require.Equal(t, gpu.VramInfo{
		AllocatedMiB:       2048,
		TotalMiB:           8192,
		UtilizationPercent: 40,
	}, visibleCard.Vram)
	require.Equal(t, gpu.GpuInfo{UtilizationPercent: 55}, visibleCard.Gpu)

	require.Equal(t, reservedUUID, reservedCard.Id)
	require.Equal(t, gpu.VramInfo{}, reservedCard.Vram)
	require.Equal(t, gpu.GpuInfo{}, reservedCard.Gpu)
}

// NVML runtime stats are enrichment: an NVML fault must not hide the card
// (or the whole list); the card is reported from hex data without stats.
func TestListLocalGpuCardsDegradesOnNvmlFaults(t *testing.T) {
	restoreGpuSeams(t)

	gpuUUID := "GPU-55555555-5555-5555-5555-555555555555"
	hexGpusMap := map[string]gpu.GpuFromHex{
		"0000:01:00.0": {
			Id:         gpuUUID,
			Name:       "NVIDIA A100",
			Type:       gpu.ResourceTypePgpu,
			PciAddress: "0000:01:00.0",
			Status:     gpu.GpuStatusIdle,
		},
	}

	getNodeGpusMap = func(nodeName string) (map[string]gpu.GpuFromHex, error) {
		return hexGpusMap, nil
	}

	t.Run("device handle fault reports card without runtime stats", func(t *testing.T) {
		deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
			return nil, nvml.ERROR_UNKNOWN
		}

		cards, err := (&helper{node: "node-1"}).listLocalGpuCards()

		require.NoError(t, err)
		require.Len(t, cards, 1)
		require.Equal(t, gpuUUID, cards[0].Id)
		require.Equal(t, gpu.VramInfo{}, cards[0].Vram)
		require.Equal(t, gpu.GpuInfo{}, cards[0].Gpu)
	})

	t.Run("memory info fault degrades vram stats only", func(t *testing.T) {
		deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
			return &nvmlmock.Device{
				GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
					return nvml.Memory{}, nvml.ERROR_UNKNOWN
				},
				GetUtilizationRatesFunc: func() (nvml.Utilization, nvml.Return) {
					return nvml.Utilization{Gpu: 55, Memory: 40}, nvml.SUCCESS
				},
			}, nvml.SUCCESS
		}

		cards, err := (&helper{node: "node-1"}).listLocalGpuCards()

		require.NoError(t, err)
		require.Len(t, cards, 1)
		require.Equal(t, gpu.VramInfo{UtilizationPercent: 40}, cards[0].Vram)
		require.Equal(t, gpu.GpuInfo{UtilizationPercent: 55}, cards[0].Gpu)
	})

	t.Run("utilization rates fault degrades utilization stats only", func(t *testing.T) {
		deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
			return &nvmlmock.Device{
				GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
					return nvml.Memory{Used: 2048 * bytesPerMiB, Total: 8192 * bytesPerMiB}, nvml.SUCCESS
				},
				GetUtilizationRatesFunc: func() (nvml.Utilization, nvml.Return) {
					return nvml.Utilization{}, nvml.ERROR_UNKNOWN
				},
			}, nvml.SUCCESS
		}

		cards, err := (&helper{node: "node-1"}).listLocalGpuCards()

		require.NoError(t, err)
		require.Len(t, cards, 1)
		require.Equal(t, gpu.VramInfo{AllocatedMiB: 2048, TotalMiB: 8192}, cards[0].Vram)
		require.Equal(t, gpu.GpuInfo{}, cards[0].Gpu)
	})
}

func TestBuildLocalGpuCardVgpuProfiles(t *testing.T) {
	restoreGpuSeams(t)

	alias := "A100-4C"
	vmCountLimit := 4

	hexGpu := gpu.GpuFromHex{
		Id:         "GPU-33333333-3333-3333-3333-333333333333",
		Name:       "NVIDIA A100",
		Type:       gpu.ResourceTypeSriovVgpu,
		PciAddress: "0000:03:00.0",
		Status:     gpu.GpuStatusInUse,
	}

	deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
		return &nvmlmock.Device{
			GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
				return nvml.Memory{Used: 0, Total: 8192 * bytesPerMiB}, nvml.SUCCESS
			},
			GetUtilizationRatesFunc: func() (nvml.Utilization, nvml.Return) {
				return nvml.Utilization{}, nvml.SUCCESS
			},
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return []nvml.VgpuInstance{}, nvml.SUCCESS
			},
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.SUCCESS
			},
		}, nvml.SUCCESS
	}

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex) {
		require.Equal(t, hexGpu.PciAddress, gpuId)

		profile := gpu.VgpuProfileFromHex{
			Id:           1,
			Name:         "GRID A100-4C",
			VramMiB:      4096,
			Count:        4,
			Alias:        &alias,
			VmCountLimit: &vmCountLimit,
		}

		return map[uint32]gpu.VgpuProfileFromHex{1: profile},
			gpu.VgpuProfileCollectionFromHex{Sriov: &[]gpu.VgpuProfileFromHex{profile}}
	}

	h := &helper{node: "node-1"}
	card, err := h.buildLocalGpuCard(hexGpu)

	require.NoError(t, err)
	require.Equal(t, hexGpu.Id, card.Id)
	require.Len(t, card.Profiles.SriovVgpu, 1)
	require.Equal(t, gpu.VgpuProfile{
		Id:         1,
		Name:       "GRID A100-4C",
		VramMiB:    4096,
		Count:      4,
		Remaining:  nil,
		AliasName:  &alias,
		CountLimit: &vmCountLimit,
	}, card.Profiles.SriovVgpu[0])
	require.Empty(t, card.Profiles.MigBackedVgpu)
	require.NotNil(t, card.AttachedInstances)
	require.Empty(t, *card.AttachedInstances)
}

func TestListAttachedInstancesUnhandledTypes(t *testing.T) {
	instances, err := listAttachedInstances(listAttachedInstancesOpts{
		HexGpu: gpu.GpuFromHex{Type: gpu.ResourceTypeUnset},
	})
	require.NoError(t, err)
	require.Nil(t, instances)

	instances, err = listAttachedInstances(listAttachedInstancesOpts{
		HexGpu: gpu.GpuFromHex{Type: gpu.ResourceType("bogus")},
	})
	require.Error(t, err)
	require.Nil(t, instances)
}

func TestListPgpuAttachedInstances(t *testing.T) {
	restoreGpuSeams(t)

	t.Run("no allocation reports no instances", func(t *testing.T) {
		instances, err := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{Type: gpu.ResourceTypePgpu},
		})

		require.NoError(t, err)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	t.Run("zero current allocation reports no instances", func(t *testing.T) {
		instances, err := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 0, Total: 1},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	t.Run("hex lookup failure returns error", func(t *testing.T) {
		getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
			return nil, errors.New("hex_sdk failed")
		}

		instances, err := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
		})

		require.Error(t, err)
		require.Nil(t, instances)
	})

	t.Run("missing hex instance returns error", func(t *testing.T) {
		getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
			return nil, nil
		}

		instances, err := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
		})

		require.Error(t, err)
		require.Nil(t, instances)
	})

	t.Run("attached instance is reported with links", func(t *testing.T) {
		getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
			return &gpu.PgpuAttachedInstanceFromHex{Id: "vm-1", Name: "instance-1"}, nil
		}

		links := gpu.InstanceLinks{Grafana: "https://grafana.example/vm-1"}
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			require.Equal(t, "vm-1", vmId)
			return links
		}

		instances, err := listPgpuAttachedInstances(listAttachedInstancesOpts{
			DeviceMemoryUsedMiB:      2048,
			DeviceMemoryTotalMiB:     8192,
			DeviceGpuUtilizationRate: 55,
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, instances)
		require.Len(t, *instances, 1)
		require.Equal(t, gpu.AttachedInstance{
			Id:                 "vm-1",
			Name:               "instance-1",
			ProfileAlias:       nil,
			UtilizationPercent: 55,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: 2048,
				TotalMiB:     8192,
			},
			Links: links,
		}, (*instances)[0])
	})
}

func TestListVgpuAttachedInstances(t *testing.T) {
	restoreGpuSeams(t)

	profileId := uint32(5)
	alias := "A100-4C"

	newVgpuInstance := func(vmId string) *nvmlmock.VgpuInstance {
		vgpuType := &nvmlmock.VgpuTypeId{
			GetGpuInstanceProfileIdFunc: func() (uint32, nvml.Return) {
				return profileId, nvml.SUCCESS
			},
			GetFramebufferSizeFunc: func() (uint64, nvml.Return) {
				return 4096 * bytesPerMiB, nvml.SUCCESS
			},
		}

		return &nvmlmock.VgpuInstance{
			GetVmIDFunc: func() (string, nvml.VgpuVmIdType, nvml.Return) {
				return vmId, nvml.VgpuVmIdType(0), nvml.SUCCESS
			},
			GetTypeFunc: func() (nvml.VgpuTypeId, nvml.Return) {
				return vgpuType, nvml.SUCCESS
			},
			GetFbUsageFunc: func() (uint64, nvml.Return) {
				return 1024 * bytesPerMiB, nvml.SUCCESS
			},
		}
	}

	newDevice := func(instances []nvml.VgpuInstance, samples []nvml.VgpuInstanceUtilizationSample) *nvmlmock.Device {
		return &nvmlmock.Device{
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return instances, nvml.SUCCESS
			},
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, samples, nvml.SUCCESS
			},
		}
	}

	t.Run("device invisible to nvml reports no instances", func(t *testing.T) {
		instances, err := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: false,
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
		})

		require.NoError(t, err)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	t.Run("active vgpu listing failure returns error", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return nil, nvml.ERROR_UNKNOWN
			},
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.SUCCESS
			},
		}

		instances, err := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
		})

		require.Error(t, err)
		require.Nil(t, instances)
	})

	t.Run("vgpu utilization failure returns error", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return []nvml.VgpuInstance{}, nvml.SUCCESS
			},
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.ERROR_UNKNOWN
			},
		}

		instances, err := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
		})

		require.Error(t, err)
		require.Nil(t, instances)
	})

	t.Run("attached instance is reported with server name", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }

		getOpenstackServer = func(vmId string) (*servers.Server, error) {
			require.Equal(t, "vm-9", vmId)
			return &servers.Server{Name: "instance-9"}, nil
		}

		links := gpu.InstanceLinks{Grafana: "https://grafana.example/vm-9"}
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return links
		}

		device := newDevice(
			[]nvml.VgpuInstance{newVgpuInstance("vm-9")},
			[]nvml.VgpuInstanceUtilizationSample{{VgpuInstance: 7, SmUtil: [8]byte{33}}},
		)

		instances, err := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			HexProfilesMap:  map[uint32]gpu.VgpuProfileFromHex{profileId: {Alias: &alias}},
		})

		require.NoError(t, err)
		require.NotNil(t, instances)
		require.Len(t, *instances, 1)
		require.Equal(t, gpu.AttachedInstance{
			Id:                 "vm-9",
			Name:               "instance-9",
			ProfileAlias:       &alias,
			UtilizationPercent: 33,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: 1024,
				TotalMiB:     4096,
			},
			Links: links,
		}, (*instances)[0])
	})

	// The server name is enrichment: a lookup failure (e.g. the VM is being
	// torn down) must not fail the whole GPU listing.
	t.Run("server lookup failure reports instance without name", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		getOpenstackServer = func(vmId string) (*servers.Server, error) {
			return nil, errors.New("openstack unavailable")
		}

		device := newDevice([]nvml.VgpuInstance{newVgpuInstance("vm-9")}, nil)

		instances, err := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
		})

		require.NoError(t, err)
		require.Len(t, *instances, 1)
		require.Equal(t, "vm-9", (*instances)[0].Id)
		require.Empty(t, (*instances)[0].Name)
	})

	t.Run("missing server reports instance without name", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		getOpenstackServer = func(vmId string) (*servers.Server, error) {
			return nil, nil
		}

		device := newDevice([]nvml.VgpuInstance{newVgpuInstance("vm-9")}, nil)

		instances, err := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
		})

		require.NoError(t, err)
		require.Len(t, *instances, 1)
		require.Equal(t, "vm-9", (*instances)[0].Id)
		require.Empty(t, (*instances)[0].Name)
	})
}

func TestBuildVgpuInstanceUtilizationMap(t *testing.T) {
	t.Run("maps unsigned int samples by vgpu instance", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, []nvml.VgpuInstanceUtilizationSample{
					{VgpuInstance: 7, SmUtil: [8]byte{42}},
					{VgpuInstance: 9, SmUtil: [8]byte{80}},
				}, nvml.SUCCESS
			},
		}

		utilizationMap, err := buildVgpuInstanceUtilizationMap(device, "GPU-uuid")

		require.NoError(t, err)
		require.Equal(t, map[uint32]uint32{7: 42, 9: 80}, utilizationMap)
	})

	// NVML returns ERROR_NOT_FOUND when no utilization samples exist yet
	// (e.g. no vGPU has run since the last query); that is not a failure.
	t.Run("no samples reports empty map", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.ERROR_NOT_FOUND
			},
		}

		utilizationMap, err := buildVgpuInstanceUtilizationMap(device, "GPU-uuid")

		require.NoError(t, err)
		require.NotNil(t, utilizationMap)
		require.Empty(t, utilizationMap)
	})

	t.Run("utilization query failure returns error", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.ERROR_UNKNOWN
			},
		}

		utilizationMap, err := buildVgpuInstanceUtilizationMap(device, "GPU-uuid")

		require.Error(t, err)
		require.Empty(t, utilizationMap)
	})
}

func TestBuildInstanceLinksViaOpenstack(t *testing.T) {
	restoreGpuSeams(t)

	// Nova refuses to create a console for a non-ACTIVE instance (409): the
	// console link is enrichment, so the Grafana link must still be reported.
	t.Run("console failure degrades to grafana-only links", func(t *testing.T) {
		createConsole = func(vmId string) (*remoteconsoles.RemoteConsole, error) {
			return nil, errors.New("instance not active (HTTP 409)")
		}

		links := buildInstanceLinksViaOpenstack("vm-1")

		require.Contains(t, links.Grafana, "/grafana/d/PVW6vU7Wz/instance")
		require.Contains(t, links.Grafana, "var-UUID=vm-1")
		require.Empty(t, links.Console)
	})

	t.Run("console url is reported when available", func(t *testing.T) {
		createConsole = func(vmId string) (*remoteconsoles.RemoteConsole, error) {
			return &remoteconsoles.RemoteConsole{URL: "https://console.example/vnc?token=abc"}, nil
		}

		links := buildInstanceLinksViaOpenstack("vm-1")

		require.Contains(t, links.Grafana, "var-UUID=vm-1")
		require.Equal(t, "https://console.example/vnc?token=abc", links.Console)
	})
}

func TestBytesToMiB(t *testing.T) {
	require.Equal(t, 0, bytesToMiB(0))
	require.Equal(t, 1, bytesToMiB(bytesPerMiB))
	require.Equal(t, 1, bytesToMiB(2*bytesPerMiB-1))
	require.Equal(t, 8192, bytesToMiB(8192*bytesPerMiB))
}

func TestIsVgpu(t *testing.T) {
	require.False(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypeUnset}))
	require.False(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypePgpu}))
	require.True(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu}))
	require.True(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypeMigBackedVgpu}))
}

func TestToProfileCollectionComputesMigRemaining(t *testing.T) {
	migAlias := "1g.10gb"

	hexProfileCollection := gpu.VgpuProfileCollectionFromHex{
		MigBacked: &[]gpu.VgpuProfileFromHex{
			{Id: 5, Name: "MIG 1g.10gb", VramMiB: 10240, Count: 4, Alias: &migAlias},
		},
	}
	attachedInstances := &[]gpu.AttachedInstance{
		{Id: "vm-1", ProfileAlias: &migAlias},
		{Id: "vm-2", ProfileAlias: &migAlias},
	}

	collection := toProfileCollection(hexProfileCollection, attachedInstances)

	require.Empty(t, collection.SriovVgpu)
	require.Len(t, collection.MigBackedVgpu, 1)
	require.NotNil(t, collection.MigBackedVgpu[0].Remaining)
	require.Equal(t, 2, *collection.MigBackedVgpu[0].Remaining)
}

func TestCreateMigProfileRemainingMap(t *testing.T) {
	alias := "1g.10gb"

	t.Run("nil inputs yield empty map", func(t *testing.T) {
		require.Empty(t, createMigProfileRemainingMap(nil, nil))
	})

	t.Run("remaining count does not go below zero", func(t *testing.T) {
		migProfiles := &[]gpu.VgpuProfileFromHex{
			{Id: 5, Count: 1, Alias: &alias},
		}
		attachedInstances := &[]gpu.AttachedInstance{
			{Id: "vm-1", ProfileAlias: &alias},
			{Id: "vm-2", ProfileAlias: &alias},
		}

		remainingMap := createMigProfileRemainingMap(migProfiles, attachedInstances)

		require.Equal(t, map[uint32]int{5: 0}, remainingMap)
	})

	t.Run("instances with unknown alias are ignored", func(t *testing.T) {
		unknownAlias := "unknown"
		migProfiles := &[]gpu.VgpuProfileFromHex{
			{Id: 5, Count: 2, Alias: &alias},
		}
		attachedInstances := &[]gpu.AttachedInstance{
			{Id: "vm-1", ProfileAlias: &unknownAlias},
		}

		remainingMap := createMigProfileRemainingMap(migProfiles, attachedInstances)

		require.Equal(t, map[uint32]int{5: 2}, remainingMap)
	})
}
