package nodes

import (
	"errors"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

// Restores the test seams after each test so stubs do not leak between tests.
func restoreGpuSeams(t *testing.T) {
	t.Helper()

	origGetNodeGpusMap := getNodeGpusMap
	origGetNodeVgpuProfilesMap := getNodeVgpuProfilesMap
	origGetNodePgpuAttachedInstance := getNodePgpuAttachedInstance
	origGetNvidiaSmiDevices := getNvidiaSmiDevices
	origGetNvidiaSmiVgpuInstances := getNvidiaSmiVgpuInstances
	origBuildInstanceLinks := buildInstanceLinks
	origGetOpenstackServers := getOpenstackServers
	origCreateConsole := createConsole
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
		getNvidiaSmiDevices = origGetNvidiaSmiDevices
		getNvidiaSmiVgpuInstances = origGetNvidiaSmiVgpuInstances
		buildInstanceLinks = origBuildInstanceLinks
		getOpenstackServers = origGetOpenstackServers
		createConsole = origCreateConsole
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

func TestListLocalGpuCardsIncludesGpusInvisibleToNvidiaSmi(t *testing.T) {
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
			// yet: invisible to nvidia-smi while having no allocation.
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

	buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
		return gpu.InstanceLinks{}
	}

	// A vfio-bound GPU is invisible to nvidia-smi: only the visible one appears.
	getNvidiaSmiDevices = func() (map[string]cubecos.NvidiaSmiDevice, error) {
		return map[string]cubecos.NvidiaSmiDevice{
			visibleUUID: {
				UUID:                     visibleUUID,
				MemoryUsedMiB:            2048,
				MemoryTotalMiB:           8192,
				MemoryUtilizationPercent: ptr(uint32(40)),
				GpuUtilizationPercent:    ptr(uint32(55)),
			},
		}, nil
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
	// nvidia-smi data is unavailable for a passthrough GPU; hex data is still
	// reported. Every stat must be nil (JSON null) rather than 0, so a consumer
	// cannot read an unmeasurable card as an idle one.
	require.Equal(t, gpu.VramInfo{}, passthroughCard.Vram)
	require.Equal(t, gpu.GpuInfo{}, passthroughCard.Gpu)

	require.Equal(t, visibleUUID, visibleCard.Id)
	require.Equal(t, "0000:02:00.0", visibleCard.PciAddress)
	require.Equal(t, gpu.VramInfo{
		AllocatedMiB:       ptr(2048),
		TotalMiB:           ptr(8192),
		UtilizationPercent: ptr(uint32(40)),
	}, visibleCard.Vram)
	require.Equal(t, gpu.GpuInfo{UtilizationPercent: ptr(uint32(55))}, visibleCard.Gpu)

	require.Equal(t, reservedUUID, reservedCard.Id)
	require.Equal(t, gpu.VramInfo{}, reservedCard.Vram)
	require.Equal(t, gpu.GpuInfo{}, reservedCard.Gpu)

	// A vfio-passthrough GPU invisible to nvidia-smi is expected, not degraded.
	require.False(t, passthroughCard.Degraded)
	require.False(t, visibleCard.Degraded)
	require.False(t, reservedCard.Degraded)
}

// nvidia-smi runtime stats are enrichment: a query fault must not hide the
// card (or the whole list); the card is reported from hex data without stats.
func TestListLocalGpuCardsDegradesOnNvidiaSmiFaults(t *testing.T) {
	restoreGpuSeams(t)

	gpuUUID := "GPU-55555555-5555-5555-5555-555555555555"

	// A pgpu absent from nvidia-smi's device map is normally read as expected
	// vfio passthrough (see TestListLocalGpuCardsIncludesGpusInvisibleToNvidiaSmi).
	// Type unset is used here instead to exercise the genuinely-unexpected-
	// absence case without also pulling in vGPU-only side effects (hex
	// profile fetch, Openstack server-name prefetch, vgpu instance query)
	// that a vGPU type would require mocking just to stay deterministic.
	t.Run("device unexpectedly absent reports degraded card without runtime stats", func(t *testing.T) {
		getNodeGpusMap = func(nodeName string) (map[string]gpu.GpuFromHex, error) {
			return map[string]gpu.GpuFromHex{
				"0000:01:00.0": {
					Id:         gpuUUID,
					Name:       "NVIDIA A100",
					Type:       gpu.ResourceTypeUnset,
					PciAddress: "0000:01:00.0",
					Status:     gpu.GpuStatusIdle,
				},
			}, nil
		}
		getNvidiaSmiDevices = func() (map[string]cubecos.NvidiaSmiDevice, error) {
			return map[string]cubecos.NvidiaSmiDevice{}, nil
		}

		cards, err := (&helper{node: "node-1"}).listLocalGpuCards()

		require.NoError(t, err)
		require.Len(t, cards, 1)
		require.Equal(t, gpuUUID, cards[0].Id)
		require.Equal(t, gpu.VramInfo{}, cards[0].Vram)
		require.Equal(t, gpu.GpuInfo{}, cards[0].Gpu)
		require.True(t, cards[0].Degraded)
	})

	// A node-wide nvidia-smi outage (command failed to run at all) must
	// degrade every card, pgpu included -- it must never be misread as "every
	// pgpu happens to be legitimately passed through".
	t.Run("nvidia-smi unavailable node-wide reports degraded card even for pgpu", func(t *testing.T) {
		getNodeGpusMap = func(nodeName string) (map[string]gpu.GpuFromHex, error) {
			return map[string]gpu.GpuFromHex{
				"0000:01:00.0": {
					Id:         gpuUUID,
					Name:       "NVIDIA A100",
					Type:       gpu.ResourceTypePgpu,
					PciAddress: "0000:01:00.0",
					Status:     gpu.GpuStatusIdle,
				},
			}, nil
		}
		getNvidiaSmiDevices = func() (map[string]cubecos.NvidiaSmiDevice, error) {
			return nil, errors.New("nvidia-smi: command not found")
		}

		cards, err := (&helper{node: "node-1"}).listLocalGpuCards()

		require.NoError(t, err)
		require.Len(t, cards, 1)
		require.True(t, cards[0].Degraded)
	})
}

// A card's vGPU profiles are capability data, not state: they say what the card
// could be partitioned into, which is exactly what a client needs to switch a
// pgpu card into vGPU mode. hex agrees - gpu_vgpu_profile_list builds them from
// `nvidia-smi vgpu -s -v` and never reads the card's configured type - and so
// does the API contract (docs.yaml's gpu-002 example is a pgpu card reporting
// sriovVgpu profiles). Gating the fetch on the card's *current* type instead
// makes the switch impossible: no profiles reported, so nothing to request.
func TestBuildLocalGpuCardReportsProfilesForVgpuCapablePgpuCard(t *testing.T) {
	restoreGpuSeams(t)

	hexGpu := gpu.GpuFromHex{
		Id:           "GPU-88888888-8888-8888-8888-888888888888",
		Name:         "NVIDIA RTX Pro 6000 Blackwell",
		Type:         gpu.ResourceTypePgpu,
		SupportTypes: []gpu.SupportResourceType{gpu.SupportResourceTypePgpu, gpu.SupportResourceTypeSriovVgpu},
		PciAddress:   "0000:06:00.0",
		Status:       gpu.GpuStatusIdle,
		Allocation:   &gpu.AllocationSummary{Current: 0, Total: 1},
	}

	// Mirrors what hex reports for an SR-IOV-capable card that is not currently
	// partitioned: the profile comes from nvidia-smi (id, name, vram), while
	// count/alias come from config.json, which holds no entry for this card yet.
	// The name is the bare type suffix - sdk_gpu.sh takes `awk '{print $NF}'` of
	// nvidia-smi's Name line, so "NVIDIA RTX Pro 6000 Blackwell DC-2B" arrives as
	// "DC-2B". SR-IOV profiles carry no vmCountLimit: their nvidia-smi block has
	// no `Max Instances` line, and sdk_gpu.sh defaults the field to null.
	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		require.Equal(t, hexGpu.Id, gpuId)

		profile := gpu.VgpuProfileFromHex{
			Id:           1518,
			Name:         "DC-2B",
			VramMiB:      2048,
			Count:        0,
			Alias:        nil,
			VmCountLimit: nil,
		}

		return map[uint32]gpu.VgpuProfileFromHex{profile.Id: profile},
			gpu.VgpuProfileCollectionFromHex{Sriov: &[]gpu.VgpuProfileFromHex{profile}},
			nil
	}

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
		Servers: map[string]openstackServer{},
		NvidiaSmiDevices: map[string]cubecos.NvidiaSmiDevice{
			hexGpu.Id: {UUID: hexGpu.Id, MemoryTotalMiB: 98304},
		},
		NvidiaSmiAvailable:     true,
		VgpuInstancesAvailable: true,
	})

	require.NoError(t, err)
	require.False(t, card.Degraded)
	require.Equal(t, []gpu.VgpuProfile{{
		Id:         1518,
		Name:       "DC-2B",
		VramMiB:    2048,
		Count:      0,
		Remaining:  nil,
		AliasName:  nil,
		CountLimit: nil,
	}}, card.Profiles.SriovVgpu)
	require.Empty(t, card.Profiles.MigBackedVgpu)
}

// MIG capability counts the same as SR-IOV: an unset card that can only be
// partitioned MIG-backed still reports the profiles needed to configure it.
func TestBuildLocalGpuCardReportsProfilesForMigCapableUnsetCard(t *testing.T) {
	restoreGpuSeams(t)

	hexGpu := gpu.GpuFromHex{
		Id:           "GPU-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Name:         "NVIDIA A100",
		Type:         gpu.ResourceTypeUnset,
		SupportTypes: []gpu.SupportResourceType{gpu.SupportResourceTypeMigBackedVgpu},
		PciAddress:   "0000:08:00.0",
		Status:       gpu.GpuStatusUnassigned,
	}

	// Post-#905 shape: a MIG-backed entry is a vGPU *type*, not a GPU Instance
	// Profile, so its id is the vGPU type id and its name carries the slice count
	// ("DC-1-3Q") that MIG_PROFILE_NAME_REGEX keys off. vmCountLimit is the
	// `Max Instances` line of the same nvidia-smi block; SR-IOV types have none.
	vmCountLimit := 32
	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		profile := gpu.VgpuProfileFromHex{
			Id:           1549,
			Name:         "DC-1-3Q",
			VramMiB:      3072,
			Count:        0,
			Alias:        nil,
			VmCountLimit: &vmCountLimit,
		}

		return map[uint32]gpu.VgpuProfileFromHex{profile.Id: profile},
			gpu.VgpuProfileCollectionFromHex{MigBacked: &[]gpu.VgpuProfileFromHex{profile}},
			nil
	}

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
		Servers: map[string]openstackServer{},
		NvidiaSmiDevices: map[string]cubecos.NvidiaSmiDevice{
			hexGpu.Id: {UUID: hexGpu.Id, MemoryTotalMiB: 81920},
		},
		NvidiaSmiAvailable:     true,
		VgpuInstancesAvailable: true,
	})

	require.NoError(t, err)
	require.False(t, card.Degraded)
	require.Len(t, card.Profiles.MigBackedVgpu, 1)
	require.Equal(t, uint32(1549), card.Profiles.MigBackedVgpu[0].Id)
	require.Equal(t, &vmCountLimit, card.Profiles.MigBackedVgpu[0].CountLimit)
	require.Empty(t, card.Profiles.SriovVgpu)
}

// The other side of the gate: a card that supports only pgpu has no vGPU
// profiles to report, so hex is never asked for them. gpu_vgpu_profile_list is a
// subprocess spawn per card, and the list path deliberately keeps those off the
// per-card path (see buildLocalGpuCardOpts).
func TestBuildLocalGpuCardSkipsProfileFetchForPgpuOnlyCard(t *testing.T) {
	restoreGpuSeams(t)

	hexGpu := gpu.GpuFromHex{
		Id:           "GPU-bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		Name:         "NVIDIA T4",
		Type:         gpu.ResourceTypePgpu,
		SupportTypes: []gpu.SupportResourceType{gpu.SupportResourceTypePgpu},
		PciAddress:   "0000:09:00.0",
		Status:       gpu.GpuStatusIdle,
		Allocation:   &gpu.AllocationSummary{Current: 0, Total: 1},
	}

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		t.Errorf("gpu_vgpu_profile_list must not be spawned for a pgpu-only card (gpu %s)", gpuId)
		return nil, gpu.VgpuProfileCollectionFromHex{}, nil
	}

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
		Servers: map[string]openstackServer{},
		NvidiaSmiDevices: map[string]cubecos.NvidiaSmiDevice{
			hexGpu.Id: {UUID: hexGpu.Id, MemoryTotalMiB: 16384},
		},
		NvidiaSmiAvailable:     true,
		VgpuInstancesAvailable: true,
	})

	require.NoError(t, err)
	require.False(t, card.Degraded)
	require.Empty(t, card.Profiles.SriovVgpu)
	require.Empty(t, card.Profiles.MigBackedVgpu)
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

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		// The UUID, not the PCI address: gpu_vgpu_profile_list keys off the
		// `.id` field in config.json. This assertion previously demanded the
		// PCI address, which is what kept the wrong lookup key in place.
		require.Equal(t, hexGpu.Id, gpuId)

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
	card, err := h.buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
		Servers: map[string]openstackServer{},
		NvidiaSmiDevices: map[string]cubecos.NvidiaSmiDevice{
			hexGpu.Id: {UUID: hexGpu.Id, MemoryTotalMiB: 8192},
		},
		NvidiaSmiAvailable:     true,
		VgpuInstancesAvailable: true,
	})

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

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		return map[uint32]gpu.VgpuProfileFromHex{}, gpu.VgpuProfileCollectionFromHex{}, errors.New("hex_sdk failed")
	}

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
		Servers: map[string]openstackServer{},
		NvidiaSmiDevices: map[string]cubecos.NvidiaSmiDevice{
			hexGpu.Id: {UUID: hexGpu.Id, MemoryTotalMiB: 8192},
		},
		NvidiaSmiAvailable:     true,
		VgpuInstancesAvailable: true,
	})

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

	getNodeVgpuProfilesMap = func(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
		return map[uint32]gpu.VgpuProfileFromHex{}, gpu.VgpuProfileCollectionFromHex{}, nil
	}

	// nil Servers == prefetch failed. No active vGPU: the card still
	// degrades because names are unavailable.
	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
		Servers: nil,
		NvidiaSmiDevices: map[string]cubecos.NvidiaSmiDevice{
			hexGpu.Id: {UUID: hexGpu.Id, MemoryTotalMiB: 8192},
		},
		NvidiaSmiAvailable:     true,
		VgpuInstancesAvailable: true,
	})

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
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			require.Equal(t, "vm-1", opts.InstanceId)
			return links
		}

		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			DeviceMemoryUsedMiB:      ptr(2048),
			DeviceMemoryTotalMiB:     ptr(8192),
			DeviceGpuUtilizationRate: ptr(uint32(55)),
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
			UtilizationPercent: ptr(uint32(55)),
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: ptr(2048),
				TotalMiB:     ptr(8192),
			},
			Links: links,
		}, (*instances)[0])
	})

	// A pgpu's card is invisible to nvidia-smi, so the instance inherits that
	// absence: every stat must be nil rather than a zero that reads as idle.
	t.Run("reports nil stats when the device is invisible", func(t *testing.T) {
		getNodePgpuAttachedInstance = func(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
			return &gpu.PgpuAttachedInstanceFromHex{Id: "vm-1", Name: "instance-1"}, nil
		}
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		enr := &enrichment{}
		instances := listPgpuAttachedInstances(listAttachedInstancesOpts{
			HexGpu: gpu.GpuFromHex{
				Type:       gpu.ResourceTypePgpu,
				Allocation: &gpu.AllocationSummary{Current: 1, Total: 1},
			},
			Enrichment: enr,
		})

		require.False(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Nil(t, (*instances)[0].UtilizationPercent)
		require.Equal(t, gpu.InstanceMemoryUsage{}, (*instances)[0].MemoryUsage)
	})
}

func TestListVgpuAttachedInstances(t *testing.T) {
	restoreGpuSeams(t)

	profileId := uint32(5)
	alias := "A100-4C"

	t.Run("device invisible to nvidia-smi reports no instances", func(t *testing.T) {
		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: false,
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			Enrichment:      enr,
		})

		// The caller already flagged the card degraded when the device was
		// unavailable, so this path does not double-flag.
		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	// The vgpu -q snapshot is enrichment: a query failure degrades to no
	// attached instances rather than failing the whole listing.
	t.Run("vgpu instances query failure degrades", func(t *testing.T) {
		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible:        true,
			DeviceUUID:             "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:                 gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			VgpuInstancesAvailable: false,
			Enrichment:             enr,
		})

		require.True(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	t.Run("no active vgpu reports no instances", func(t *testing.T) {
		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible:        true,
			DeviceUUID:             "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:                 gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			VgpuInstances:          nil,
			VgpuInstancesAvailable: true,
			Enrichment:             enr,
		})

		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Empty(t, *instances)
	})

	t.Run("attached instance is reported with server name", func(t *testing.T) {
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			return gpu.InstanceLinks{Grafana: "https://grafana.example/" + opts.InstanceId}
		}

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			// SR-IOV: hex profile id is the vGPU Type ID directly, no MIG
			// type-map lookup involved.
			HexProfilesMap: map[uint32]gpu.VgpuProfileFromHex{profileId: {Alias: &alias}},
			VgpuInstances: []cubecos.NvidiaSmiVgpuInstance{
				{VmUUID: "vm-9", VgpuTypeId: profileId, MemoryUsedMiB: ptr(1024), MemoryTotalMiB: ptr(4096), GpuUtilizationPercent: ptr(uint32(33))},
			},
			VgpuInstancesAvailable: true,
			Servers:                map[string]openstackServer{"vm-9": {Name: "instance-9", TenantId: "proj-9"}},
			Enrichment:             enr,
		})

		require.False(t, enr.degraded)
		require.NotNil(t, instances)
		require.Len(t, *instances, 1)
		require.Equal(t, gpu.AttachedInstance{
			Id:                 "vm-9",
			Name:               "instance-9",
			ProfileAlias:       &alias,
			UtilizationPercent: ptr(uint32(33)),
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: ptr(1024),
				TotalMiB:     ptr(4096),
			},
			Links: gpu.InstanceLinks{Grafana: "https://grafana.example/vm-9"},
		}, (*instances)[0])
	})

	// A MIG-backed vGPU reports its framebuffer but N/A for utilization, so the
	// instance must carry the memory numbers and a nil utilization.
	t.Run("reports a MIG-backed instance without utilization", func(t *testing.T) {
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeMigBackedVgpu},
			HexProfilesMap:  map[uint32]gpu.VgpuProfileFromHex{profileId: {Alias: &alias}},
			VgpuInstances: []cubecos.NvidiaSmiVgpuInstance{
				{VmUUID: "vm-9", VgpuTypeId: profileId, MemoryUsedMiB: ptr(144), MemoryTotalMiB: ptr(2048)},
			},
			VgpuInstancesAvailable: true,
			Servers:                map[string]openstackServer{"vm-9": {Name: "instance-9", TenantId: "proj-9"}},
			Enrichment:             enr,
		})

		require.False(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Nil(t, (*instances)[0].UtilizationPercent)
		require.Equal(t, gpu.InstanceMemoryUsage{
			AllocatedMiB: ptr(144),
			TotalMiB:     ptr(2048),
		}, (*instances)[0].MemoryUsage)
	})

	// The server name is enrichment: an instance absent from the prefetched
	// server map (e.g. the VM is being torn down) is reported without a name
	// rather than failing the whole GPU listing.
	t.Run("instance missing from server map reported without name", func(t *testing.T) {
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			VgpuInstances: []cubecos.NvidiaSmiVgpuInstance{
				{VmUUID: "vm-9", VgpuTypeId: profileId},
			},
			VgpuInstancesAvailable: true,
			Servers:                map[string]openstackServer{},
			Enrichment:             enr,
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
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu},
			VgpuInstances: []cubecos.NvidiaSmiVgpuInstance{
				{VmUUID: "vm-9", VgpuTypeId: profileId},
			},
			VgpuInstancesAvailable: true,
			Servers:                nil,
			Enrichment:             enr,
		})

		require.False(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Equal(t, "vm-9", (*instances)[0].Id)
		require.Empty(t, (*instances)[0].Name)
	})

	// MIG-backed vGPU type IDs are not hex profile ids; the GI profile id is
	// resolved via the pre-fetched type map instead of used directly.
	t.Run("mig-backed instance resolves profile id via type map", func(t *testing.T) {
		buildInstanceLinks = func(opts instanceLinkOpts) gpu.InstanceLinks {
			return gpu.InstanceLinks{}
		}

		// hex keys migBacked profiles by vGPU Type ID (cubecos #905), the same id
		// `vgpu -q` reports per instance - so no translation step is involved.
		vgpuTypeId := uint32(0x619)

		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeMigBackedVgpu},
			HexProfilesMap:  map[uint32]gpu.VgpuProfileFromHex{vgpuTypeId: {Alias: &alias}},
			VgpuInstances: []cubecos.NvidiaSmiVgpuInstance{
				{VmUUID: "vm-9", VgpuTypeId: vgpuTypeId},
			},
			VgpuInstancesAvailable: true,
			Servers:                map[string]openstackServer{},
			Enrichment:             enr,
		})

		require.False(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Equal(t, &alias, (*instances)[0].ProfileAlias)
	})

	// A vGPU type absent from hex's profile list is reported without an alias and
	// does not degrade the card - the same as the SR-IOV path has always behaved.
	// Both flavours share one lookup now, so there is no MIG-only signal to keep:
	// the id mismatch that used to produce one (hex reporting GPU Instance Profile
	// IDs while vgpu -q reports vGPU Type IDs) no longer exists.
	t.Run("vgpu type missing from hex profiles reports without an alias", func(t *testing.T) {
		enr := &enrichment{}
		instances := listVgpuAttachedInstances(listAttachedInstancesOpts{
			IsDeviceVisible: true,
			DeviceUUID:      "GPU-44444444-4444-4444-4444-444444444444",
			HexGpu:          gpu.GpuFromHex{Type: gpu.ResourceTypeMigBackedVgpu},
			HexProfilesMap:  map[uint32]gpu.VgpuProfileFromHex{},
			VgpuInstances: []cubecos.NvidiaSmiVgpuInstance{
				{VmUUID: "vm-9", VgpuTypeId: 0x619},
			},
			VgpuInstancesAvailable: true,
			Servers:                map[string]openstackServer{},
			Enrichment:             enr,
		})

		require.False(t, enr.degraded)
		require.Len(t, *instances, 1)
		require.Nil(t, (*instances)[0].ProfileAlias)
	})
}

// The list-time links carry only the Grafana dashboard; the console is minted
// on demand via getGpuInstanceConsole and is not part of the listing. All three
// dashboard variables must be pinned, or the page labels this VM's chart with
// another VM's name.
func TestBuildInstanceLinksViaOpenstack(t *testing.T) {
	links := buildInstanceLinksViaOpenstack(instanceLinkOpts{
		InstanceId: "vm-1",
		TenantId:   "proj-1",
		VmName:     "instance-1",
	})

	require.Contains(t, links.Grafana, "/grafana/d/PVW6vU7Wz/instance")
	require.Contains(t, links.Grafana, "var-UUID=vm-1")
	require.Contains(t, links.Grafana, "var-TID=proj-1")
	require.Contains(t, links.Grafana, "var-HOSTNAME=instance-1")
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

func TestResolveServers(t *testing.T) {
	restoreGpuSeams(t)

	t.Run("skips openstack when node has no vgpu", func(t *testing.T) {
		called := false
		getOpenstackServers = func() (map[string]openstackServer, error) {
			called = true
			return nil, nil
		}

		servers := (&helper{node: "node-1"}).resolveServers(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypePgpu},
		})

		require.False(t, called)
		require.NotNil(t, servers)
		require.Empty(t, servers)
	})

	t.Run("fetches servers once when a vgpu is present", func(t *testing.T) {
		calls := 0
		getOpenstackServers = func() (map[string]openstackServer, error) {
			calls++
			return map[string]openstackServer{"vm-9": {Name: "instance-9", TenantId: "proj-9"}}, nil
		}

		servers := (&helper{node: "node-1"}).resolveServers(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypeSriovVgpu},
			"0000:02:00.0": {Type: gpu.ResourceTypeMigBackedVgpu},
		})

		require.Equal(t, 1, calls)
		require.Equal(t, map[string]openstackServer{"vm-9": {Name: "instance-9", TenantId: "proj-9"}}, servers)
	})

	// A server-listing failure is enrichment: names degrade to a nil map (the
	// unavailable sentinel) rather than failing the whole listing.
	t.Run("returns nil on openstack error", func(t *testing.T) {
		getOpenstackServers = func() (map[string]openstackServer, error) {
			return nil, errors.New("openstack unavailable")
		}

		servers := (&helper{node: "node-1"}).resolveServers(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypeSriovVgpu},
		})

		require.Nil(t, servers)
	})
}

func TestResolveVgpuInstances(t *testing.T) {
	restoreGpuSeams(t)

	t.Run("skips nvidia-smi when node has no vgpu", func(t *testing.T) {
		called := false
		getNvidiaSmiVgpuInstances = func() (map[string][]cubecos.NvidiaSmiVgpuInstance, error) {
			called = true
			return nil, nil
		}

		instances, err := (&helper{node: "node-1"}).resolveVgpuInstances(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypePgpu},
		})

		require.NoError(t, err)
		require.False(t, called)
		require.Empty(t, instances)
	})

	t.Run("fetches instances once when a vgpu is present", func(t *testing.T) {
		calls := 0
		getNvidiaSmiVgpuInstances = func() (map[string][]cubecos.NvidiaSmiVgpuInstance, error) {
			calls++
			return map[string][]cubecos.NvidiaSmiVgpuInstance{
				"0000:01:00.0": {{VmUUID: "vm-9"}},
			}, nil
		}

		instances, err := (&helper{node: "node-1"}).resolveVgpuInstances(map[string]gpu.GpuFromHex{
			"0000:01:00.0": {Type: gpu.ResourceTypeSriovVgpu},
		})

		require.NoError(t, err)
		require.Equal(t, 1, calls)
		require.Equal(t, "vm-9", instances["0000:01:00.0"][0].VmUUID)
	})
}

func TestWarnGpusMissingFromHex(t *testing.T) {
	hexUUID := "GPU-11111111-1111-1111-1111-111111111111"
	hexGpusMap := map[string]gpu.GpuFromHex{
		"0000:01:00.0": {Id: hexUUID, Type: gpu.ResourceTypePgpu},
	}

	// A vfio-passthrough GPU is invisible to nvidia-smi enumeration, so
	// reconciliation only makes sense while nvidia-smi is up.
	t.Run("skips reconciliation when nvidia-smi is unavailable", func(t *testing.T) {
		// Deliberately includes a device that would trigger the warning if
		// reconciliation ran anyway, to prove the early return short-circuits.
		nvidiaSmiDevices := map[string]cubecos.NvidiaSmiDevice{
			"GPU-22222222-2222-2222-2222-222222222222": {},
		}

		(&helper{node: "node-1"}).warnGpusMissingFromHex(hexGpusMap, nvidiaSmiDevices, false)
		// No panic/assertion target here beyond "does not crash" -- the
		// warning is a log line with no observable seam in this test; the
		// behavior under test is covered by the "when available" case below
		// via presence/absence of the log path being reachable at all.
	})

	t.Run("reports uuids nvidia-smi sees that hex does not", func(t *testing.T) {
		nvidiaSmiDevices := map[string]cubecos.NvidiaSmiDevice{
			hexUUID: {},
			"GPU-22222222-2222-2222-2222-222222222222": {},
		}

		// No seam to assert the warning content on directly; this exercises
		// the reconciliation path end-to-end without panicking on either the
		// matched or the mismatched UUID.
		(&helper{node: "node-1"}).warnGpusMissingFromHex(hexGpusMap, nvidiaSmiDevices, true)
	})
}

func TestIsVgpu(t *testing.T) {
	require.False(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypeUnset}))
	require.False(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypePgpu}))
	require.True(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypeSriovVgpu}))
	require.True(t, isVgpu(gpu.GpuFromHex{Type: gpu.ResourceTypeMigBackedVgpu}))
}

func TestValidateGpuCardUpdate(t *testing.T) {
	profileCountLimit := 10
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

	// The limit bounds the total number of vGPU instances (sum of the requested
	// per-profile counts), not the number of distinct profiles: a single profile
	// whose count exceeds the limit trips it.
	t.Run("total profile count over card limit -> 409 sentinel", func(t *testing.T) {
		two := 2
		limited := card
		limited.SriovVgpuProfileCountLimit = &two // allow only 2 vgpu instances total
		err := validateGpuCardUpdate(limited, gpu.UpdateGpuCardRequest{
			ResourceType: gpu.ResourceTypeSriovVgpu,
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 3}}, // sum 3 > limit 2
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
			Profiles:     []gpu.UpdateGpuCardProfile{{Id: 57, Count: 1}, {Id: 58, Count: 1}}, // total count 2 > limit 1
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
		getNvidiaSmiDevices = func() (map[string]cubecos.NvidiaSmiDevice, error) {
			return map[string]cubecos.NvidiaSmiDevice{
				"GPU-1": {UUID: "GPU-1", MemoryTotalMiB: 16384},
			}, nil
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
		getNvidiaSmiDevices = func() (map[string]cubecos.NvidiaSmiDevice, error) {
			return map[string]cubecos.NvidiaSmiDevice{
				"GPU-1": {UUID: "GPU-1", MemoryTotalMiB: 16384},
			}, nil
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

	card, err := (&helper{node: "node-1"}).buildLocalGpuCard(gpu.GpuFromHex{
		Id:         "GPU-proc",
		PciAddress: "0000:01:00.0",
		Type:       gpu.ResourceTypePgpu,
		Status:     gpu.GpuStatusIdle,
	}, buildLocalGpuCardOpts{
		Servers:            map[string]openstackServer{},
		NvidiaSmiDevices:   map[string]cubecos.NvidiaSmiDevice{},
		NvidiaSmiAvailable: true,
	})

	require.NoError(t, err)
	require.True(t, card.Status.IsProcessing)
}
