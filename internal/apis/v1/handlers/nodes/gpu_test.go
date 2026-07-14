package nodes

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	nvmlmock "github.com/NVIDIA/go-nvml/pkg/nvml/mock"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
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
	origDeviceGetCount := deviceGetCount
	origDeviceGetHandleByIndex := deviceGetHandleByIndex
	origBuildInstanceLinks := buildInstanceLinks
	origGetOpenstackServerNames := getOpenstackServerNames
	origCreateConsole := createConsole
	origVgpuInstanceId := vgpuInstanceId
	origIsNvmlAvailable := isNvmlAvailable
	origGetNodeGpuById := getNodeGpuById
	origUpdateNodeGpuCardViaHex := updateNodeGpuCardViaHex
	origIsGpuUpdating := isGpuUpdating
	origUpsertUpdatingGpuReq := upsertUpdatingGpuReq
	origDeleteUpdatingGpuReq := deleteUpdatingGpuReq

	// Defaults that avoid MongoDB in tests that build/list cards.
	isGpuUpdating = func(h *helper, gpuId string) bool { return false }
	upsertUpdatingGpuReq = func(h *helper, gpuId string) error { return nil }
	deleteUpdatingGpuReq = func(h *helper, gpuId string) error { return nil }

	t.Cleanup(func() {
		getNodeGpusMap = origGetNodeGpusMap
		getNodeVgpuProfilesMap = origGetNodeVgpuProfilesMap
		getNodePgpuAttachedInstance = origGetNodePgpuAttachedInstance
		deviceGetHandleByUUID = origDeviceGetHandleByUUID
		deviceGetCount = origDeviceGetCount
		deviceGetHandleByIndex = origDeviceGetHandleByIndex
		buildInstanceLinks = origBuildInstanceLinks
		getOpenstackServerNames = origGetOpenstackServerNames
		createConsole = origCreateConsole
		vgpuInstanceId = origVgpuInstanceId
		isNvmlAvailable = origIsNvmlAvailable
		getNodeGpuById = origGetNodeGpuById
		updateNodeGpuCardViaHex = origUpdateNodeGpuCardViaHex
		isGpuUpdating = origIsGpuUpdating
		upsertUpdatingGpuReq = origUpsertUpdatingGpuReq
		deleteUpdatingGpuReq = origDeleteUpdatingGpuReq
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

	// A vfio-passthrough GPU invisible to NVML is expected, not degraded.
	require.False(t, passthroughCard.Degraded)
	require.False(t, visibleCard.Degraded)
	require.False(t, reservedCard.Degraded)
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

	// An unexpected handle fault on a GPU that NVML should have seen leaves the
	// card without trustworthy capacity, so it is flagged degraded.
	t.Run("device handle fault reports degraded card without runtime stats", func(t *testing.T) {
		deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
			return nil, nvml.ERROR_UNKNOWN
		}

		cards, err := (&helper{node: "node-1"}).listLocalGpuCards()

		require.NoError(t, err)
		require.Len(t, cards, 1)
		require.Equal(t, gpuUUID, cards[0].Id)
		require.Equal(t, gpu.VramInfo{}, cards[0].Vram)
		require.Equal(t, gpu.GpuInfo{}, cards[0].Gpu)
		require.True(t, cards[0].Degraded)
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
		// Lost memory info is lost capacity: the card is flagged degraded.
		require.True(t, cards[0].Degraded)
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
		// Lost utilization is untrustworthy stats: the card is flagged degraded.
		require.True(t, cards[0].Degraded)
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

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
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
			gpu.VgpuProfileCollectionFromHex{Sriov: &[]gpu.VgpuProfileFromHex{profile}},
			nil
	}

	h := &helper{node: "node-1"}
	card, err := h.buildLocalGpuCard(hexGpu, map[string]string{})

	require.NoError(t, err)
	require.False(t, card.Degraded)
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

// A failed hex vgpu-profile fetch leaves the card's advertised capacity
// untrustworthy, so the card is flagged degraded rather than reporting an empty
// profile set as if it were real.
func TestBuildLocalGpuCardDegradesOnProfileFetchFailure(t *testing.T) {
	restoreGpuSeams(t)

	hexGpu := gpu.GpuFromHex{
		Id:         "GPU-66666666-6666-6666-6666-666666666666",
		Name:       "NVIDIA A100",
		Type:       gpu.ResourceTypeSriovVgpu,
		PciAddress: "0000:04:00.0",
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
		}, nvml.SUCCESS
	}

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		return map[uint32]gpu.VgpuProfileFromHex{}, gpu.VgpuProfileCollectionFromHex{}, errors.New("hex_sdk failed")
	}

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, map[string]string{})

	require.NoError(t, err)
	require.True(t, card.Degraded)
}

// A nil server-name map means the Openstack prefetch failed, so a vGPU card's
// attached-instance names are all unavailable: the card is flagged degraded even
// when it currently has no attached instances.
func TestBuildLocalGpuCardDegradesOnServerPrefetchFailure(t *testing.T) {
	restoreGpuSeams(t)

	hexGpu := gpu.GpuFromHex{
		Id:         "GPU-77777777-7777-7777-7777-777777777777",
		Name:       "NVIDIA A100",
		Type:       gpu.ResourceTypeSriovVgpu,
		PciAddress: "0000:05:00.0",
		Status:     gpu.GpuStatusIdle,
	}

	deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
		return &nvmlmock.Device{
			GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
				return nvml.Memory{Used: 0, Total: 8192 * bytesPerMiB}, nvml.SUCCESS
			},
			GetUtilizationRatesFunc: func() (nvml.Utilization, nvml.Return) {
				return nvml.Utilization{}, nvml.SUCCESS
			},
			// No active vGPU: the card still degrades because names are unavailable.
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return []nvml.VgpuInstance{}, nvml.SUCCESS
			},
		}, nvml.SUCCESS
	}

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		return map[uint32]gpu.VgpuProfileFromHex{}, gpu.VgpuProfileCollectionFromHex{}, nil
	}

	// nil serverNames == prefetch failed.
	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, nil)

	require.NoError(t, err)
	require.True(t, card.Degraded)
}

func TestListAttachedInstancesUnhandledTypes(t *testing.T) {
	enr := &enrichment{}

	instances, err := listAttachedInstances(listAttachedInstancesOpts{
		HexGpu:     gpu.GpuFromHex{Type: gpu.ResourceTypeUnset},
		Enrichment: enr,
	})
	require.NoError(t, err)
	require.False(t, enr.degraded)
	require.Nil(t, instances)

	instances, err = listAttachedInstances(listAttachedInstancesOpts{
		HexGpu:     gpu.GpuFromHex{Type: gpu.ResourceType("bogus")},
		Enrichment: enr,
	})
	require.Error(t, err)
	require.False(t, enr.degraded)
	require.Nil(t, instances)
}

func TestListPgpuAttachedInstances(t *testing.T) {
	restoreGpuSeams(t)

	t.Run("no allocation reports no instances", func(t *testing.T) {
		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu:     gpu.GpuFromHex{Type: gpu.ResourceTypePgpu},
			Enrichment: enr,
		})

		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	t.Run("zero current allocation reports no instances", func(t *testing.T) {
		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 0, Total: 1},
			},
			Enrichment: enr,
		})

		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	// A failed hex attached-instance lookup on an allocated pgpu degrades to no
	// attached instance and flags the card degraded rather than failing the whole
	// node listing.
	t.Run("hex lookup failure degrades", func(t *testing.T) {
		getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
			return nil, errors.New("hex_sdk failed")
		}

		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
			Enrichment: enr,
		})

		require.True(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	// A transient hex read skew (allocation already counted while the attached
	// instance is not yet reported) must degrade to no attached instance rather
	// than failing the whole node listing.
	t.Run("missing hex instance degrades to no instances", func(t *testing.T) {
		getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
			return nil, nil
		}

		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
			Enrichment: enr,
		})

		// A transient hex read skew is a benign out-of-sync read, not a lookup
		// failure: the pgpu path stays soft without flagging the card.
		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
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

		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			DeviceMemoryUsedMiB:      2048,
			DeviceMemoryTotalMiB:     8192,
			DeviceGpuUtilizationRate: 55,
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
			Enrichment: enr,
		})

		require.False(t, enr.degraded)
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
		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: false,
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			Enrichment:      enr,
		})

		// The caller already flagged the card degraded when the handle failed, so
		// this path does not double-flag.
		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	// The active vGPU list is NVML enrichment (e.g. GPU_IS_LOST during a reset):
	// it degrades to no attached instances rather than failing the whole listing.
	t.Run("active vgpu listing failure degrades", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return nil, nvml.ERROR_GPU_IS_LOST
			},
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.SUCCESS
			},
		}

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			Enrichment:      enr,
		})

		require.True(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	// No active vGPU short-circuits before the utilization query, so a failing
	// query there is never even reached.
	t.Run("no active vgpu reports no instances", func(t *testing.T) {
		device := &nvmlmock.Device{
			GetActiveVgpusFunc: func() ([]nvml.VgpuInstance, nvml.Return) {
				return []nvml.VgpuInstance{}, nvml.SUCCESS
			},
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_UNSIGNED_INT, nil, nvml.ERROR_UNKNOWN
			},
		}

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			Enrichment:      enr,
		})

		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	// A vGPU instance whose per-instance NVML call fails (e.g. the VM was torn
	// down between GetActiveVgpus and GetVmID, leaving a stale handle) is skipped
	// rather than failing the whole node listing.
	t.Run("stale instance handle is skipped", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		staleInstance := &nvmlmock.VgpuInstance{
			GetVmIDFunc: func() (string, nvml.VgpuVmIdType, nvml.Return) {
				return "", nvml.VgpuVmIdType(0), nvml.ERROR_NOT_FOUND
			},
		}

		device := newDevice(
			[]nvml.VgpuInstance{staleInstance, newVgpuInstance("vm-9")},
			nil,
		)

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			HexProfilesMap:  map[uint32]gpu.VgpuProfileFromHex{profileId: {Alias: &alias}},
			ServerNames:     map[string]string{"vm-9": "instance-9"},
			Enrichment:      enr,
		})

		// A skipped instance must flag the card degraded so the missing instance
		// is not read as a real detach.
		require.True(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Equal(t, "vm-9", (*instances)[0].Id)
	})

	t.Run("attached instance is reported with server name", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }

		links := gpu.InstanceLinks{Grafana: "https://grafana.example/vm-9"}
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return links
		}

		device := newDevice(
			[]nvml.VgpuInstance{newVgpuInstance("vm-9")},
			[]nvml.VgpuInstanceUtilizationSample{{VgpuInstance: 7, SmUtil: [8]byte{33}}},
		)

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			HexProfilesMap:  map[uint32]gpu.VgpuProfileFromHex{profileId: {Alias: &alias}},
			ServerNames:     map[string]string{"vm-9": "instance-9"},
			Enrichment:      enr,
		})

		require.False(t, enr.degraded)
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

	// The server name is enrichment: an instance absent from the prefetched
	// server map (e.g. the VM is being torn down) is reported without a name
	// rather than failing the whole GPU listing.
	t.Run("instance missing from server map reported without name", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		device := newDevice([]nvml.VgpuInstance{newVgpuInstance("vm-9")}, nil)

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			ServerNames:     map[string]string{},
			Enrichment:      enr,
		})

		// A missing server name from a non-nil map is soft enrichment, not a lost
		// instance: the card is not degraded.
		require.False(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Equal(t, "vm-9", (*instances)[0].Id)
		require.Empty(t, (*instances)[0].Name)
	})

	// A nil server-name map (prefetch failed) is handled at the card level by
	// buildLocalGpuCard, not here: this layer only reports the instance without a
	// name and does not itself degrade. See TestBuildLocalGpuCardDegradesOnServerPrefetchFailure.
	t.Run("nil server map does not degrade at this layer", func(t *testing.T) {
		vgpuInstanceId = func(instance nvml.VgpuInstance) uint32 { return 7 }
		buildInstanceLinks = func(vmId string) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		device := newDevice([]nvml.VgpuInstance{newVgpuInstance("vm-9")}, nil)

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			Device:          device,
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			ServerNames:     nil,
			Enrichment:      enr,
		})

		require.False(t, enr.degraded)
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

		enr := &enrichment{}
		utilizationMap := buildVgpuInstanceUtilizationMap(device, "GPU-uuid", enr)

		require.False(t, enr.degraded)
		require.Equal(t, map[uint32]uint32{7: 42, 9: 80}, utilizationMap)
	})

	// Any non-SUCCESS return (no samples yet, unsupported hardware, GPU lost, or
	// unknown) would leave every attached instance reporting 0% utilization, which
	// reads as idle rather than unknown: degrade to an empty map instead of failing
	// the whole listing.
	t.Run("utilization query failure degrades to empty map", func(t *testing.T) {
		for _, ret := range []nvml.Return{nvml.ERROR_NOT_FOUND, nvml.ERROR_NOT_SUPPORTED, nvml.ERROR_GPU_IS_LOST, nvml.ERROR_UNKNOWN} {
			device := &nvmlmock.Device{
				GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
					return nvml.VALUE_TYPE_UNSIGNED_INT, nil, ret
				},
			}

			enr := &enrichment{}
			utilizationMap := buildVgpuInstanceUtilizationMap(device, "GPU-uuid", enr)

			require.True(t, enr.degraded)
			require.NotNil(t, utilizationMap)
			require.Empty(t, utilizationMap)
		}
	})

	// A double-typed sample carries the raw bytes of an IEEE-754 double and must
	// be reinterpreted, not read as a little-endian integer.
	t.Run("maps double samples by vgpu instance", func(t *testing.T) {
		var smUtil [8]byte
		binary.LittleEndian.PutUint64(smUtil[:], math.Float64bits(33.0))

		device := &nvmlmock.Device{
			GetVgpuUtilizationFunc: func(v uint64) (nvml.ValueType, []nvml.VgpuInstanceUtilizationSample, nvml.Return) {
				return nvml.VALUE_TYPE_DOUBLE, []nvml.VgpuInstanceUtilizationSample{
					{VgpuInstance: 7, SmUtil: smUtil},
				}, nvml.SUCCESS
			},
		}

		enr := &enrichment{}
		utilizationMap := buildVgpuInstanceUtilizationMap(device, "GPU-uuid", enr)

		require.False(t, enr.degraded)
		require.Equal(t, map[uint32]uint32{7: 33}, utilizationMap)
	})
}

// The list-time links carry only the Grafana dashboard; the console is minted
// on demand via getGpuInstanceConsole and is not part of the listing.
func TestBuildInstanceLinksViaOpenstack(t *testing.T) {
	links := buildInstanceLinksViaOpenstack("vm-1")

	require.Contains(t, links.Grafana, "/grafana/d/PVW6vU7Wz/instance")
	require.Contains(t, links.Grafana, "var-UUID=vm-1")
}

func TestGetGpuInstanceConsole(t *testing.T) {
	restoreGpuSeams(t)

	t.Run("console url is returned when minted", func(t *testing.T) {
		createConsole = func(vmId string) (*remoteconsoles.RemoteConsole, error) {
			require.Equal(t, "vm-1", vmId)
			return &remoteconsoles.RemoteConsole{URL: "https://console.example/vnc?token=abc"}, nil
		}

		console, err := (&helper{instanceId: "vm-1"}).getGpuInstanceConsole()

		require.NoError(t, err)
		require.Equal(t, "https://console.example/vnc?token=abc", console.Console)
	})

	// Nova refuses to create a console for a non-ACTIVE instance (409): the
	// on-demand endpoint surfaces that as an error to the caller.
	t.Run("console creation failure returns error", func(t *testing.T) {
		createConsole = func(vmId string) (*remoteconsoles.RemoteConsole, error) {
			return nil, errors.New("instance not active (HTTP 409)")
		}

		console, err := (&helper{instanceId: "vm-1"}).getGpuInstanceConsole()

		require.Error(t, err)
		require.Nil(t, console)
	})

	t.Run("nil console returns error", func(t *testing.T) {
		createConsole = func(vmId string) (*remoteconsoles.RemoteConsole, error) {
			return nil, nil
		}

		console, err := (&helper{instanceId: "vm-1"}).getGpuInstanceConsole()

		require.Error(t, err)
		require.Nil(t, console)
	})
}

func TestResolveServerNames(t *testing.T) {
	restoreGpuSeams(t)

	t.Run("skips openstack when node has no vgpu", func(t *testing.T) {
		called := false
		getOpenstackServerNames = func() (map[string]string, error) {
			called = true
			return nil, nil
		}

		names := (&helper{node: "node-1"}).resolveServerNames(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypePgpu},
		})

		require.False(t, called)
		require.NotNil(t, names)
		require.Empty(t, names)
	})

	t.Run("fetches names once when a vgpu is present", func(t *testing.T) {
		calls := 0
		getOpenstackServerNames = func() (map[string]string, error) {
			calls++
			return map[string]string{"vm-9": "instance-9"}, nil
		}

		names := (&helper{node: "node-1"}).resolveServerNames(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypeSriovVgpu},
			"0000:02:00.0": {Type: gpu.ResourceTypeMigBackedVgpu},
		})

		require.Equal(t, 1, calls)
		require.Equal(t, map[string]string{"vm-9": "instance-9"}, names)
	})

	// A server-listing failure is enrichment: names degrade to a nil map (the
	// unavailable sentinel) rather than failing the whole listing.
	t.Run("returns nil on openstack error", func(t *testing.T) {
		getOpenstackServerNames = func() (map[string]string, error) {
			return nil, errors.New("openstack unavailable")
		}

		names := (&helper{node: "node-1"}).resolveServerNames(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypeSriovVgpu},
		})

		require.Nil(t, names)
	})
}

func TestWarnGpusMissingFromHex(t *testing.T) {
	restoreGpuSeams(t)

	hexUUID := "GPU-11111111-1111-1111-1111-111111111111"
	hexGpusMap := map[string]gpu.GpuFromHex{
		"0000:01:00.0": {Id: hexUUID, Type: gpu.ResourceTypePgpu},
	}

	// A vfio-passthrough GPU is invisible to NVML enumeration, so reconciliation
	// only makes sense while NVML is up.
	t.Run("skips enumeration when nvml is unavailable", func(t *testing.T) {
		isNvmlAvailable = func() bool { return false }
		called := false
		deviceGetCount = func() (int, nvml.Return) {
			called = true
			return 0, nvml.SUCCESS
		}

		(&helper{node: "node-1"}).warnGpusMissingFromHex(hexGpusMap)

		require.False(t, called)
	})

	t.Run("enumerates every nvml device when available", func(t *testing.T) {
		isNvmlAvailable = func() bool { return true }
		deviceGetCount = func() (int, nvml.Return) { return 2, nvml.SUCCESS }

		var indices []int
		deviceGetHandleByIndex = func(i int) (nvml.Device, nvml.Return) {
			indices = append(indices, i)

			uuid := hexUUID
			if i == 1 {
				// A GPU visible to NVML but absent from hex (the case the warning
				// is meant to surface).
				uuid = "GPU-22222222-2222-2222-2222-222222222222"
			}

			return &nvmlmock.Device{
				GetUUIDFunc: func() (string, nvml.Return) { return uuid, nvml.SUCCESS },
			}, nvml.SUCCESS
		}

		(&helper{node: "node-1"}).warnGpusMissingFromHex(hexGpusMap)

		require.Equal(t, []int{0, 1}, indices)
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

func TestValidateGpuCardUpdate(t *testing.T) {
	profileCountLimit := 2
	vmCountLimit := 4

	card := gpu.GpuFromHex{
		Id:                         "GPU-1",
		PciAddress:                 "0000:01:00.0",
		SupportTypes:               []gpu.SupportResourceType{gpu.SupportResourceTypePgpu, gpu.SupportResourceTypeSriovVgpu},
		Status:                     gpu.GpuStatusIdle,
		SriovVgpuProfileCountLimit: &profileCountLimit,
	}

	// Only mig-backed profiles carry a vmCountLimit.
	// Two profiles: id 57 = 2048 MiB each (vm limit 4), id 58 = 4096 MiB each (no vm limit).
	profilesMap := map[uint32]gpu.VgpuProfileFromHex{
		57: {Id: 57, VramMiB: 2048, VmCountLimit: &vmCountLimit},
		58: {Id: 58, VramMiB: 4096},
	}
	totalVramMiB := 16384

	t.Run("unsupported resource type -> 409 sentinel", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypeMigBackedVgpu}, nil, 0)
		require.ErrorIs(t, err, gpu.ErrUnsupportedType)
	})

	t.Run("profiles on a non-vgpu type -> 400 sentinel", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypePgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 1}},
		}, profilesMap, totalVramMiB)
		require.ErrorIs(t, err, gpu.ErrProfilesNotAllowed)
	})

	t.Run("gpu in use by status -> 409 sentinel", func(t *testing.T) {
		inUse := card
		inUse.Status = gpu.GpuStatusInUse
		err := validateGpuCardUpdate(inUse, gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypePgpu}, nil, 0)
		require.ErrorIs(t, err, gpu.ErrGpuInUse)
	})

	t.Run("gpu in use by allocation -> 409 sentinel", func(t *testing.T) {
		inUse := card
		inUse.Allocation = &gpu.AllocationSummary{Current: 1, Total: 1}
		err := validateGpuCardUpdate(inUse, gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypePgpu}, nil, 0)
		require.ErrorIs(t, err, gpu.ErrGpuInUse)
	})

	t.Run("unknown profile id -> 400 sentinel", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 999, Count: 1}},
		}, profilesMap, totalVramMiB)
		require.ErrorIs(t, err, gpu.ErrProfileNotFound)
	})

	t.Run("distinct profile count over card limit -> 409 sentinel", func(t *testing.T) {
		three := 1
		limited := card
		limited.SriovVgpuProfileCountLimit = &three // allow only 1 distinct profile
		err := validateGpuCardUpdate(limited, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 1}, {Id: 58, Count: 1}},
		}, profilesMap, totalVramMiB)
		require.ErrorIs(t, err, gpu.ErrExceedProfileCountLimit)
	})

	// The profile-count limit is SR-IOV only; MIG-backed vGPU ignores it.
	t.Run("mig-backed vgpu ignores the profile count limit", func(t *testing.T) {
		one := 1
		migCard := card
		migCard.SupportTypes = []gpu.SupportResourceType{gpu.SupportResourceTypePgpu, gpu.SupportResourceTypeMigBackedVgpu}
		migCard.SriovVgpuProfileCountLimit = &one // would trip for SR-IOV, but must be ignored here
		err := validateGpuCardUpdate(migCard, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeMigBackedVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 1}, {Id: 58, Count: 1}}, // 2 distinct > limit 1
		}, profilesMap, totalVramMiB)
		require.NoError(t, err)
	})

	// The per-profile vmCountLimit is MIG-backed only.
	t.Run("mig-backed per-profile count over vmCountLimit -> 409 sentinel", func(t *testing.T) {
		migCard := card
		migCard.SupportTypes = []gpu.SupportResourceType{gpu.SupportResourceTypePgpu, gpu.SupportResourceTypeMigBackedVgpu}
		err := validateGpuCardUpdate(migCard, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeMigBackedVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 5}}, // vmCountLimit is 4; 2048*5 = 10240 <= 16384 so vram is not what trips
		}, profilesMap, totalVramMiB)
		require.ErrorIs(t, err, gpu.ErrExceedProfileCountLimit)
	})

	t.Run("sriov vgpu ignores the per-profile vmCountLimit", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 5}}, // vmCountLimit is 4, but ignored for sriov
		}, profilesMap, totalVramMiB)
		require.NoError(t, err)
	})

	// The VRAM limit applies to MIG-backed vGPU only.
	t.Run("mig-backed total requested vram over device total -> 409 sentinel", func(t *testing.T) {
		migCard := card
		migCard.SupportTypes = []gpu.SupportResourceType{gpu.SupportResourceTypePgpu, gpu.SupportResourceTypeMigBackedVgpu}
		err := validateGpuCardUpdate(migCard, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeMigBackedVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 58, Count: 5}}, // 4096*5 = 20480 > 16384
		}, profilesMap, 16384)
		require.ErrorIs(t, err, gpu.ErrExceedVramLimit)
	})

	// SR-IOV profiles are fixed partitions; their total VRAM is not checked against device memory.
	t.Run("sriov vgpu ignores the vram limit", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 58, Count: 5}}, // 20480 > 16384, but ignored for sriov
		}, profilesMap, 16384)
		require.NoError(t, err)
	})

	t.Run("valid pgpu update passes", func(t *testing.T) {
		require.NoError(t, validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypePgpu}, nil, 0))
	})

	// TODO: What?
	t.Run("valid sriov vgpu update within all limits passes", func(t *testing.T) {
		require.NoError(t, validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 2}, {Id: 58, Count: 1}}, // 2048*2+4096 = 8192
		}, profilesMap, totalVramMiB))
	})

	t.Run("non-positive profile count -> 400 sentinel", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 0}},
		}, profilesMap, totalVramMiB)
		require.ErrorIs(t, err, gpu.ErrInvalidProfileCount)
	})

	t.Run("duplicate profile id -> 400 sentinel", func(t *testing.T) {
		err := validateGpuCardUpdate(card, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 1}, {Id: 57, Count: 1}}, // same id twice
		}, profilesMap, totalVramMiB)
		require.ErrorIs(t, err, gpu.ErrDuplicateProfile)
	})
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

	// TODO: but this should not happen. We should log warning.
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

	// TODO: This should not happen. We should log warning.
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

func TestUpdateLocalGpuCard(t *testing.T) {
	restoreGpuSeams(t)

	card := gpu.GpuFromHex{
		Id:           "GPU-1",
		PciAddress:   "0000:01:00.0",
		SupportTypes: []gpu.SupportResourceType{gpu.SupportResourceTypePgpu, gpu.SupportResourceTypeSriovVgpu},
		Status:       gpu.GpuStatusIdle,
	}

	t.Run("valid pgpu update calls hex and clears the record", func(t *testing.T) {
		getNodeGpuById = func(nodeName, gpuId string) (gpu.GpuFromHex, error) { return card, nil }

		var hexCalled, upserted, deleted bool
		upsertUpdatingGpuReq = func(h *helper, gpuId string) error { upserted = true; return nil }
		deleteUpdatingGpuReq = func(h *helper, gpuId string) error { deleted = true; return nil }
		updateNodeGpuCardViaHex = func(gpuId string, req gpu.UpdateGpuCardRequest) error {
			hexCalled = true
			require.Equal(t, "GPU-1", gpuId)
			require.Equal(t, gpu.ResourceTypePgpu, req.ResourceType)
			return nil
		}

		h := &helper{node: "node-1", gpuId: "GPU-1", gpuCardReq: gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypePgpu,
		}}

		require.NoError(t, h.updateLocalGpuCard())
		require.True(t, upserted)
		require.True(t, hexCalled)
		require.True(t, deleted)
	})

	t.Run("valid sriov vgpu update gathers profiles + vram then calls hex", func(t *testing.T) {
		vmLimit := 4
		getNodeGpuById = func(nodeName, gpuId string) (gpu.GpuFromHex, error) { return card, nil }
		getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
			require.Equal(t, card.Id, gpuId)
			return map[uint32]gpu.VgpuProfileFromHex{
				57: {Id: 57, VramMiB: 2048, VmCountLimit: &vmLimit},
			}, gpu.VgpuProfileCollectionFromHex{}, nil
		}
		deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
			require.Equal(t, "GPU-1", uuid)
			return &nvmlmock.Device{
				GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
					return nvml.Memory{Total: 16384 * bytesPerMiB}, nvml.SUCCESS
				},
			}, nvml.SUCCESS
		}

		var hexCalled bool
		updateNodeGpuCardViaHex = func(gpuId string, req gpu.UpdateGpuCardRequest) error { hexCalled = true; return nil }

		h := &helper{node: "node-1", gpuId: "GPU-1", gpuCardReq: gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 2}}, // 2048*2 = 4096 <= 16384
		}}

		require.NoError(t, h.updateLocalGpuCard())
		require.True(t, hexCalled)
	})

	t.Run("hex failure still clears the record", func(t *testing.T) {
		getNodeGpuById = func(nodeName, gpuId string) (gpu.GpuFromHex, error) { return card, nil }

		var deleted bool
		deleteUpdatingGpuReq = func(h *helper, gpuId string) error { deleted = true; return nil }
		updateNodeGpuCardViaHex = func(gpuId string, req gpu.UpdateGpuCardRequest) error {
			return errors.New("hex_config failed")
		}

		h := &helper{node: "node-1", gpuId: "GPU-1", gpuCardReq: gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypePgpu}}

		require.Error(t, h.updateLocalGpuCard())
		require.True(t, deleted)
	})

	t.Run("validation failure skips hex and record", func(t *testing.T) {
		getNodeGpuById = func(nodeName, gpuId string) (gpu.GpuFromHex, error) { return card, nil }

		upsertUpdatingGpuReq = func(h *helper, gpuId string) error { t.Fatal("must not upsert"); return nil }
		updateNodeGpuCardViaHex = func(gpuId string, req gpu.UpdateGpuCardRequest) error { t.Fatal("must not call hex"); return nil }

		h := &helper{node: "node-1", gpuId: "GPU-1", gpuCardReq: gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypeMigBackedVgpu}}

		err := h.updateLocalGpuCard()
		require.ErrorIs(t, err, gpu.ErrUnsupportedType)
	})

	t.Run("gpu not found is surfaced", func(t *testing.T) {
		getNodeGpuById = func(nodeName, gpuId string) (gpu.GpuFromHex, error) {
			return gpu.GpuFromHex{}, gpu.ErrGpuNotFound
		}

		h := &helper{node: "node-1", gpuId: "missing", gpuCardReq: gpu.UpdateGpuCardRequest{ResourceType: gpu.ResourceTypePgpu}}

		require.ErrorIs(t, h.updateLocalGpuCard(), gpu.ErrGpuNotFound)
	})

	t.Run("node-side profile-list failure surfaces as a non-4xx infra error, not a bad request", func(t *testing.T) {
		getNodeGpuById = func(nodeName, gpuId string) (gpu.GpuFromHex, error) { return card, nil }
		getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
			return nil, gpu.VgpuProfileCollectionFromHex{}, errors.New("hex_sdk profile list failed")
		}
		deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) {
			return &nvmlmock.Device{
				GetMemoryInfoFunc: func() (nvml.Memory, nvml.Return) {
					return nvml.Memory{Total: 16384 * bytesPerMiB}, nvml.SUCCESS
				},
			}, nvml.SUCCESS
		}

		hexCalled := false
		updateNodeGpuCardViaHex = func(gpuId string, req gpu.UpdateGpuCardRequest) error { hexCalled = true; return nil }

		h := &helper{node: "node-1", gpuId: "GPU-1", gpuCardReq: gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 1}},
		}}

		err := h.updateLocalGpuCard()

		require.Error(t, err)
		require.NotErrorIs(t, err, gpu.ErrGpuNotFound)
		require.NotErrorIs(t, err, gpu.ErrUnsupportedType)
		require.NotErrorIs(t, err, gpu.ErrProfilesNotAllowed)
		require.NotErrorIs(t, err, gpu.ErrProfileNotFound)
		require.NotErrorIs(t, err, gpu.ErrInvalidProfileCount)
		require.NotErrorIs(t, err, gpu.ErrExceedProfileCountLimit)
		require.NotErrorIs(t, err, gpu.ErrExceedVramLimit)
		require.NotErrorIs(t, err, gpu.ErrGpuInUse)
		require.False(t, hexCalled)
	})
}

func TestBuildLocalGpuCardReportsIsProcessing(t *testing.T) {
	restoreGpuSeams(t)

	isGpuUpdating = func(h *helper, gpuId string) bool { return gpuId == "GPU-proc" }
	deviceGetHandleByUUID = func(uuid string) (nvml.Device, nvml.Return) { return nil, nvml.ERROR_NOT_FOUND }

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(gpu.GpuFromHex{
		Id:         "GPU-proc",
		PciAddress: "0000:01:00.0",
		Type:       gpu.ResourceTypePgpu,
		Status:     gpu.GpuStatusIdle,
	}, nil)

	require.NoError(t, err)
	require.True(t, card.Status.IsProcessing)
}
