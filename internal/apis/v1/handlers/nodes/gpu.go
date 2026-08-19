package nodes

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// remoteConsoleMicroversion is the first Nova microversion that serves
// POST /servers/{id}/remote-consoles.
const remoteConsoleMicroversion = "2.6"

// Seams for unit tests: hex_sdk CLI, nvidia-smi, and Openstack are
// unavailable there.
var (
	getNodeGpusMap              = cubecos.GetNodeGpusMap
	getNodeVgpuProfilesMap      = cubecos.GetNodeVgpuProfilesMap
	getNodePgpuAttachedInstance = cubecos.GetNodePgpuAttachedInstance
	getNvidiaSmiDevices         = cubecos.GetNvidiaSmiDevices
	getNvidiaSmiVgpuInstances   = cubecos.GetNvidiaSmiVgpuInstances
	buildInstanceLinks          = buildInstanceLinksViaOpenstack
	buildGpuCardLinks           = buildGpuCardLinksViaGrafana
	getOpenstackServers         = getOpenstackServersViaHelper
	createConsole               = createConsoleViaOpenstack
	getNodeGpuById              = cubecos.GetNodeGpuById
	updateNodeGpuCardViaHex     = cubecos.UpdateNodeGpuCard
	isGpuUpdating               = isGpuUpdatingViaMongo
	upsertUpdatingGpuReq        = upsertUpdatingGpuReqViaMongo
	deleteUpdatingGpuReq        = deleteUpdatingGpuReqViaMongo
)

// openstackServer is the per-VM enrichment an attached instance needs from
// Openstack: its display name, and the project that owns it -- the latter only
// so the Grafana deep link can pin the dashboard's tenant variable to the right
// project (see grafana.InstanceVgpuWorkloadHistoryLink).
type openstackServer struct {
	Name     string
	TenantId string
}

// buildLocalGpuCardOpts carries data fetched once per listLocalGpuCards
// request (Openstack server names, nvidia-smi's device and vGPU-instance
// snapshots) into each card's build. nvidia-smi is a subprocess spawn, not an
// in-process library call, so it is invoked once per request here rather than
// once per GPU on the node.
type buildLocalGpuCardOpts struct {
	Servers map[string]openstackServer
	// NvidiaSmiDevices/NvidiaSmiAvailable come from one `nvidia-smi -q` call.
	// NvidiaSmiAvailable is false only when that call itself could not be run;
	// a device simply absent from NvidiaSmiDevices (e.g. vfio-pci passthrough)
	// is a normal, expected outcome even when NvidiaSmiAvailable is true.
	NvidiaSmiDevices   map[string]cubecos.NvidiaSmiDevice
	NvidiaSmiAvailable bool
	// VgpuInstances/VgpuInstancesAvailable come from one `nvidia-smi vgpu -q`
	// call already filtered to this card's PCI address.
	VgpuInstances          []cubecos.NvidiaSmiVgpuInstance
	VgpuInstancesAvailable bool
}

type listAttachedInstancesOpts struct {
	IsDeviceVisible bool
	DeviceUUID      string
	// The Device* stats are nil when nvidia-smi could not report them; a pgpu's
	// attached instance is reported with the card's own numbers, so it inherits
	// their absence rather than substituting a zero.
	DeviceMemoryUsedMiB      *int
	DeviceMemoryTotalMiB     *int
	DeviceGpuUtilizationRate *uint32
	NodeName                 string
	HexGpu                   gpu.GpuFromHex
	HexProfilesMap           map[uint32]gpu.VgpuProfileFromHex
	// VgpuInstances/VgpuInstancesAvailable: see buildLocalGpuCardOpts.
	VgpuInstances          []cubecos.NvidiaSmiVgpuInstance
	VgpuInstancesAvailable bool
	// Servers maps Openstack server id to that VM's name and owning project,
	// prefetched once per request so an attached instance does not cost a
	// GetServer round trip each. A nil map means the prefetch failed outright, so
	// a missing entry is a real enrichment loss (degrade); a non-nil map that
	// simply lacks an id means that one VM is absent (soft, reported without a
	// name).
	Servers map[string]openstackServer
	// Enrichment accumulates the card's degraded state as attached-instance
	// enrichment is gathered.
	Enrichment *enrichment
}

// enrichment accumulates the degraded state of a single GPU card while its
// nvidia-smi, hex and Openstack enrichment is gathered. degrade flags the
// card and logs why; coupling flag and log in one call keeps a "(degraded)"
// warning from ever drifting apart from the flag it is supposed to
// accompany. Gaps that are merely incomplete but not misleading (an instance
// reported without a name, a transient hex read skew) log directly via
// log.Warnf without flagging.
type enrichment struct {
	degraded bool
}

// degrade flags the card degraded and logs why. Use it when an omission could be
// misread as a real state change (a dropped instance looks like a detach, an
// empty profile set looks like real capacity).
func (e *enrichment) degrade(format string, args ...any) {
	e.degraded = true
	log.Errorf(format, args...)
}

func (h *helper) listNodeGpuCards() ([]gpu.GpuCard, error) {
	if nodes.IsLocal(h.node) {
		return h.listLocalGpuCards()
	}
	return h.listRemoteGpuCards()
}

func (h *helper) listLocalGpuCards() ([]gpu.GpuCard, error) {
	hexGpusMap, err := getNodeGpusMap(h.node)
	if err != nil {
		return nil, err
	}

	// Resolve vGPU instance names in a single Openstack round trip rather than
	// one GetServer per attached instance. A nil map signals the prefetch failed,
	// which degrades any vGPU card that has attached instances needing a name.
	serversById := h.resolveServers(hexGpusMap)

	nvidiaSmiDevices, nvidiaSmiErr := getNvidiaSmiDevices()
	if nvidiaSmiErr != nil {
		log.Errorf("gpu(%s): nvidia-smi is unavailable on node %s: %v; gpu runtime stats will be reported as degraded", h.reqId, h.node, nvidiaSmiErr)
		nvidiaSmiDevices = map[string]cubecos.NvidiaSmiDevice{}
	}

	vgpuInstancesByPci, vgpuErr := h.resolveVgpuInstances(hexGpusMap)
	if vgpuErr != nil {
		log.Errorf("gpu(%s): failed to query vgpu instances on node %s: %v; vgpu attached instances will be reported as degraded", h.reqId, h.node, vgpuErr)
	}

	gpuCards := []gpu.GpuCard{}

	// Iterate over hex GPUs instead of nvidia-smi devices: a GPU passed
	// through to a VM is invisible to nvidia-smi, but must still be reported.
	for _, pciAddress := range slices.Sorted(maps.Keys(hexGpusMap)) {
		hexGpu := hexGpusMap[pciAddress]

		gpuCard, err := h.buildLocalGpuCard(hexGpu, buildLocalGpuCardOpts{
			Servers:                serversById,
			NvidiaSmiDevices:       nvidiaSmiDevices,
			NvidiaSmiAvailable:     nvidiaSmiErr == nil,
			VgpuInstances:          vgpuInstancesByPci[hexGpu.PciAddress],
			VgpuInstancesAvailable: vgpuErr == nil,
		})
		if err != nil {
			return nil, err
		}

		gpuCards = append(gpuCards, gpuCard)
	}

	// Hex is the source of truth for the response, but a GPU visible to
	// nvidia-smi yet missing from hex points at a stale hex inventory
	// (hot-add, cache lag): warn so the mismatch is not entirely silent.
	h.warnGpusMissingFromHex(hexGpusMap, nvidiaSmiDevices, nvidiaSmiErr == nil)

	return gpuCards, nil
}

// resolveServers prefetches the Openstack servers (id -> name + owning project)
// when the node has any vGPU, whose attached-instance enrichment comes from
// Openstack. On a lookup failure it returns nil (instances reported without a
// name) rather than failing the listing; the nil map lets callers mark affected
// cards degraded.
func (h *helper) resolveServers(hexGpusMap map[string]gpu.GpuFromHex) map[string]openstackServer {
	needsServers := false
	for _, hexGpu := range hexGpusMap {
		if isVgpu(hexGpu) {
			needsServers = true
			break
		}
	}
	if !needsServers {
		return map[string]openstackServer{}
	}

	serversById, err := getOpenstackServers()
	if err != nil {
		log.Warnf("gpu(%s): failed to list Openstack servers for name resolution on node %s: %v; reporting instances without names", h.reqId, h.node, err)
		return nil
	}

	return serversById
}

// resolveVgpuInstances fetches active vGPU instances once for the whole node
// when hex reports at least one vGPU-type GPU, skipping the nvidia-smi call
// entirely on pgpu-only nodes (mirrors resolveServers's lazy Openstack
// prefetch).
func (h *helper) resolveVgpuInstances(hexGpusMap map[string]gpu.GpuFromHex) (map[string][]cubecos.NvidiaSmiVgpuInstance, error) {
	needsVgpu := false
	for _, hexGpu := range hexGpusMap {
		if isVgpu(hexGpu) {
			needsVgpu = true
			break
		}
	}
	if !needsVgpu {
		return map[string][]cubecos.NvidiaSmiVgpuInstance{}, nil
	}

	return getNvidiaSmiVgpuInstances()
}

func (h *helper) buildLocalGpuCard(hexGpu gpu.GpuFromHex, opts buildLocalGpuCardOpts) (gpu.GpuCard, error) {
	// All four stay nil unless nvidia-smi actually reported them: an absent
	// device (pgpu bound to vfio-pci) has no stats, and a MIG-enabled card has
	// no utilization. nil is reported as JSON null so a consumer can tell
	// "cannot be measured" from a real 0.
	var memoryUsedMiB, memoryTotalMiB *int
	var memoryUtilizationPercent, gpuUtilizationPercent *uint32
	enr := &enrichment{}

	// Relies on hex reporting the GPU id as nvidia-smi's own UUID.
	device, isDeviceVisible := opts.NvidiaSmiDevices[hexGpu.Id]

	// nvidia-smi runtime stats are enrichment on top of the hex inventory:
	// when they are unavailable the card is still reported, just without
	// stats.
	switch {
	case isDeviceVisible:
		// A returned device record always carries both framebuffer numbers;
		// utilization may still be nil (MIG).
		usedMiB, totalMiB := device.MemoryUsedMiB, device.MemoryTotalMiB
		memoryUsedMiB, memoryTotalMiB = &usedMiB, &totalMiB
		memoryUtilizationPercent = device.MemoryUtilizationPercent
		gpuUtilizationPercent = device.GpuUtilizationPercent
	case !opts.NvidiaSmiAvailable:
		// A node-wide nvidia-smi outage must degrade every card, pgpu included:
		// this case is checked before the pgpu-passthrough case below so an
		// outage is never misread as "every pgpu happens to be passed through".
		enr.degrade("nvidiasmi: nvidia-smi is not available on this node; gpu %s reported from hex without runtime stats or attachments (capacity is degraded)", hexGpu.Id)
	case hexGpu.Type == gpu.ResourceTypePgpu:
		// Expected: a pgpu bound to vfio for passthrough (attached to a VM or
		// reserved for one) is invisible to nvidia-smi. Only a pgpu can
		// disappear this way, so any other type missing (checked with
		// nvidia-smi confirmed available, from the case above) is genuinely
		// unexpected and falls through to the warning below.
		log.Debugf("nvidiasmi: pgpu %s not visible to nvidia-smi (expected for vfio passthrough); reporting from hex without runtime stats", hexGpu.Id)
	default:
		// nvidia-smi is available and was expected to see this device but did
		// not report it, so the card is reported without runtime stats or
		// attached instances and its capacity is untrustworthy: flag it degraded.
		enr.degrade("nvidiasmi: device %s not reported by nvidia-smi; reporting card as degraded", hexGpu.Id)
	}

	hexProfilesMap := map[uint32]gpu.VgpuProfileFromHex{}
	hexProfileCollection := gpu.VgpuProfileCollectionFromHex{}

	// Profiles are capability data, not state: hex builds them from nvidia-smi
	// and never reads the configured type, and a client needs them to switch a
	// pgpu/unset card into vGPU mode. Gating on the current type alone made that
	// a dead end - no profiles reported, so nothing to request.
	//
	// The current type still counts on its own, as a second chance rather than a
	// safeguard: supportTypes and the profile list both come from
	// `nvidia-smi vgpu -s -v`, but from two hex calls made at different moments
	// (gpu_device_list when the card list was built, gpu_vgpu_profile_list here).
	// A probe that failed for the first can succeed for the second, so a card
	// that is demonstrably running vGPU still gets asked.
	//
	// It buys no degraded flag. gpu_vgpu_profile_list sends nvidia-smi's stderr
	// to /dev/null and never checks its exit status, so a failed probe returns
	// {"sriov":[],"migBacked":[]} with exit 0 and the fetch below sees no error.
	// Empty profile lists therefore never mean "degraded" - they mean either "no
	// vGPU types" or "the probe failed", and the two are indistinguishable here.
	if supportsVgpu(hexGpu) || isVgpu(hexGpu) {
		var profErr error
		// Must be the GPU UUID, not the PCI address: gpu_vgpu_profile_list looks
		// the card up in /etc/cube/cos/gpu/config.json by its `.id` field, which
		// holds the UUID. A PCI address silently matches nothing there, so every
		// profile comes back with count 0 and a nil alias while nvidia-smi still
		// fills in the rest - a wrong answer that looks like a valid one. The
		// update path below already passes card.Id; this call was the odd one out.
		hexProfilesMap, hexProfileCollection, profErr = getNodeVgpuProfilesMap(hexGpu.Id)
		if profErr != nil {
			// Profiles drive the card's advertised vGPU capacity; a failed hex
			// profile fetch leaves that capacity untrustworthy, so flag degraded
			// rather than reporting an empty profile set as if it were real.
			enr.degrade("gpu: failed to get vgpu profiles for gpu %s on node %s: %v; reporting card as degraded", hexGpu.Id, h.node, profErr)
		}
	}

	// A nil server map means the getOpenstackServers call failed, so
	// attached-instance name enrichment is unavailable for the whole node: flag
	// the card degraded.
	if opts.Servers == nil {
		enr.degrade("gpu: Openstack server prefetch failed on node %s; gpu %s reported without instance names (degraded)", h.node, hexGpu.Id)
	}

	attachedInstances, err := listAttachedInstances(listAttachedInstancesOpts{
		IsDeviceVisible:          isDeviceVisible,
		DeviceUUID:               hexGpu.Id,
		DeviceMemoryUsedMiB:      memoryUsedMiB,
		DeviceMemoryTotalMiB:     memoryTotalMiB,
		DeviceGpuUtilizationRate: gpuUtilizationPercent,
		NodeName:                 h.node,
		HexGpu:                   hexGpu,
		HexProfilesMap:           hexProfilesMap,
		VgpuInstances:            opts.VgpuInstances,
		VgpuInstancesAvailable:   opts.VgpuInstancesAvailable,
		Servers:                  opts.Servers,
		Enrichment:               enr,
	})
	if err != nil {
		return gpu.GpuCard{}, err
	}

	profileCollection := toProfileCollection(hexProfileCollection, attachedInstances)

	gpuCard := gpu.GpuCard{
		Id:                   hexGpu.Id,
		Name:                 hexGpu.Name,
		ResourceType:         hexGpu.Type,
		SupportResourceTypes: hexGpu.SupportTypes,
		Vram: gpu.VramInfo{
			AllocatedMiB:       memoryUsedMiB,
			TotalMiB:           memoryTotalMiB,
			UtilizationPercent: memoryUtilizationPercent,
		},
		Gpu: gpu.GpuInfo{
			UtilizationPercent: gpuUtilizationPercent,
		},
		PciAddress: hexGpu.PciAddress,
		Status: gpu.GpuStatusInfo{
			Current:      hexGpu.Status,
			IsProcessing: isGpuUpdating(h, hexGpu.Id),
		},
		AllocationSummary:          hexGpu.Allocation,
		SriovVgpuProfileCountLimit: hexGpu.SriovVgpuProfileCountLimit,
		Profiles:                   profileCollection,
		AttachedInstances:          attachedInstances,
		Links:                      buildGpuCardLinks(h.node, hexGpu.PciAddress),
		Degraded:                   enr.degraded,
	}

	return gpuCard, nil
}

// warnGpusMissingFromHex warns for any nvidia-smi device that hex did not
// report. Hex drives the response, so an nvidia-smi-visible GPU absent from
// hex would otherwise silently disappear; the old code hard-errored on this.
// Runs only when nvidia-smi is available (a vfio-passthrough GPU is invisible
// to nvidia-smi and legitimately absent, so it never triggers a warning).
func (h *helper) warnGpusMissingFromHex(hexGpusMap map[string]gpu.GpuFromHex, nvidiaSmiDevices map[string]cubecos.NvidiaSmiDevice, nvidiaSmiAvailable bool) {
	if !nvidiaSmiAvailable {
		return
	}

	hexUUIDs := map[string]struct{}{}
	for _, hexGpu := range hexGpusMap {
		hexUUIDs[hexGpu.Id] = struct{}{}
	}

	for uuid := range nvidiaSmiDevices {
		if _, ok := hexUUIDs[uuid]; !ok {
			log.Warnf("gpu: nvidia-smi reports device %s that is absent from the hex inventory on node %s; hex may be stale (hot-add, cache lag)", uuid, h.node)
		}
	}
}

func (h *helper) listRemoteGpuCards() ([]gpu.GpuCard, error) {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("gpu(%s): failed to get node %s: %v", h.reqId, h.node, err)
		return nil, err
	}

	resp, err := h.http.R().
		SetResult(&gpuCardsResp{}).
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.ListGpuCardsUrl())
	if err != nil {
		log.Errorf("gpu(%s): failed to get GPU cards from remote node %s: %v", h.reqId, h.node, err)
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error response from node %s listing GPU cards: %s", h.node, string(resp.Body()))
	}

	result := resp.Result().(*gpuCardsResp)
	return result.Data, nil
}

func isVgpu(hexGpu gpu.GpuFromHex) bool {
	return isVgpuType(hexGpu.Type)
}

// supportsVgpu reports whether the card can be put into a vGPU mode, whatever
// mode it is in now. Use it for capability data (the profile list); use isVgpu
// for runtime data that only exists while the card actually runs vGPU (attached
// instances, Openstack server-name resolution).
func supportsVgpu(hexGpu gpu.GpuFromHex) bool {
	return isSupportedType(hexGpu.SupportTypes, gpu.ResourceTypeSriovVgpu) ||
		isSupportedType(hexGpu.SupportTypes, gpu.ResourceTypeMigBackedVgpu)
}

func isVgpuType(t gpu.ResourceType) bool {
	return t == gpu.ResourceTypeSriovVgpu || t == gpu.ResourceTypeMigBackedVgpu
}

func isSupportedType(supportTypes []gpu.SupportResourceType, t gpu.ResourceType) bool {
	for _, st := range supportTypes {
		if string(st) == string(t) {
			return true
		}
	}
	return false
}

// validateGpuCardUpdate checks a resource-type change against the card's real
// capabilities. profilesMap (keyed by profile Id) and totalVramMiB (from
// nvidia-smi) are only consulted for vGPU types with profiles. Returns a
// sentinel-wrapped error the handler maps to a status.
func validateGpuCardUpdate(
	card gpu.GpuFromHex,
	req gpu.UpdateGpuCardRequest,
	profilesMap map[uint32]gpu.VgpuProfileFromHex,
	totalVramMiB int,
) error {
	if !isSupportedType(card.SupportTypes, req.ResourceType) {
		return fmt.Errorf("gpu card %s does not support '%s' resource type: %w", card.Id, req.ResourceType, gpu.ErrUnsupportedType)
	}

	if !isVgpuType(req.ResourceType) && len(req.Profiles) > 0 {
		return fmt.Errorf("profiles are not allowed for resource type %s: %w", req.ResourceType, gpu.ErrProfilesNotAllowed)
	}

	if card.Status == gpu.GpuStatusInUse || (card.Allocation != nil && card.Allocation.Current > 0) {
		return fmt.Errorf("gpu %s is in use: %w", card.Id, gpu.ErrGpuInUse)
	}

	if !isVgpuType(req.ResourceType) {
		return nil
	}

	// Reject duplicate profile ids up front.
	distinctIds := map[uint32]struct{}{}
	for _, reqProfile := range req.Profiles {
		if _, dup := distinctIds[reqProfile.Id]; dup {
			return fmt.Errorf("gpu %s: duplicate vgpu profile %d: %w", card.Id, reqProfile.Id, gpu.ErrDuplicateProfile)
		}
		distinctIds[reqProfile.Id] = struct{}{}
	}

	totalVramReqMiB := 0
	totalProfileCount := 0
	for _, reqProfile := range req.Profiles {
		if reqProfile.Count <= 0 {
			return fmt.Errorf("gpu %s: profile %d count %d must be positive: %w",
				card.Id, reqProfile.Id, reqProfile.Count, gpu.ErrInvalidProfileCount)
		}

		profile, ok := profilesMap[reqProfile.Id]
		if !ok {
			return fmt.Errorf("gpu %s: vgpu profile %d not found: %w", card.Id, reqProfile.Id, gpu.ErrProfileNotFound)
		}

		// The per-profile count limit only applies to MIG-backed vGPU; only its
		// profiles carry a vm count limit.
		if req.ResourceType == gpu.ResourceTypeMigBackedVgpu &&
			profile.VmCountLimit != nil && reqProfile.Count > *profile.VmCountLimit {
			return fmt.Errorf("gpu %s: profile %d count %d exceeds vm count limit %d: %w",
				card.Id, reqProfile.Id, reqProfile.Count, *profile.VmCountLimit, gpu.ErrExceedProfileCountLimit)
		}

		totalVramReqMiB += int(profile.VramMiB) * reqProfile.Count
		totalProfileCount += reqProfile.Count
	}

	// The profile-count limit only applies to SR-IOV vGPU; MIG-backed vGPU has no such limit.
	// It bounds the total number of vGPU instances (the sum of the requested per-profile
	// counts), not the number of distinct profiles.
	if req.ResourceType == gpu.ResourceTypeSriovVgpu &&
		card.SriovVgpuProfileCountLimit != nil && totalProfileCount > *card.SriovVgpuProfileCountLimit {
		return fmt.Errorf("gpu %s: total profile count %d exceeds profile count limit %d: %w",
			card.Id, totalProfileCount, *card.SriovVgpuProfileCountLimit, gpu.ErrExceedProfileCountLimit)
	}

	// The VRAM limit only applies to MIG-backed vGPU; SR-IOV profiles are fixed
	// partitions constrained by the profile-count limit instead.
	if req.ResourceType == gpu.ResourceTypeMigBackedVgpu && totalVramReqMiB > totalVramMiB {
		return fmt.Errorf("gpu %s: requested vram %d MiB exceeds device total %d MiB: %w",
			card.Id, totalVramReqMiB, totalVramMiB, gpu.ErrExceedVramLimit)
	}

	return nil
}

// listAttachedInstances returns the attached instances, flagging the card via
// opts.Enrichment when nvidia-smi/hex enrichment that should have populated
// them could not be obtained. An error is reserved for integrity failures (an
// unhandled type) that must fail the whole listing.
func listAttachedInstances(opts listAttachedInstancesOpts) (*[]gpu.AttachedInstance, error) {
	hexGpu := opts.HexGpu

	switch opts.HexGpu.Type {
	case gpu.ResourceTypeUnset:
		return nil, nil
	case gpu.ResourceTypePgpu:
		return listPgpuAttachedInstances(opts), nil
	case gpu.ResourceTypeSriovVgpu, gpu.ResourceTypeMigBackedVgpu:
		return listVgpuAttachedInstances(opts), nil
	default:
		return nil, fmt.Errorf("gpu: unhandled gpu type %s when listing attached instances for gpu %s", hexGpu.Type, hexGpu.Id)
	}
}

// listPgpuAttachedInstances resolves a pgpu's attached instance from hex. A
// failed hex attached-instance lookup degrades (the allocation is real but the
// instance detail is unavailable, so no instance must not be read as absent); a
// transient allocation/instance read skew is soft.
func listPgpuAttachedInstances(opts listAttachedInstancesOpts) *[]gpu.AttachedInstance {
	deviceMemoryUsedMiB,
		deviceMemoryTotalMiB,
		deviceGpuUtilizationRate,
		nodeName,
		hexGpu :=
		opts.DeviceMemoryUsedMiB,
		opts.DeviceMemoryTotalMiB,
		opts.DeviceGpuUtilizationRate,
		opts.NodeName,
		opts.HexGpu

	attachedInstances := []gpu.AttachedInstance{}

	if hexGpu.Allocation == nil || hexGpu.Allocation.Current == 0 {
		return &attachedInstances
	}

	attachedInstance, err := getNodePgpuAttachedInstance(hexGpu.PciAddress)
	if err != nil {
		opts.Enrichment.degrade("gpu: failed to get attached instance for pgpu %s on node %s: %v; reporting no attached instance (degraded)", hexGpu.PciAddress, nodeName, err)
		return &attachedInstances
	}

	if attachedInstance == nil {
		// The allocation count and the attached-instance lookup are two separate
		// hex reads: during an attach/detach they can be observed out of sync
		// (allocation already 1 while the instance is not yet reported). This is a
		// benign transient skew, not a lost enrichment, so it logs without degrading.
		log.Warnf("gpu: hex reports allocation for pgpu %s on node %s but no attached instance yet (transient read skew); reporting no attached instance", hexGpu.PciAddress, nodeName)
		return &attachedInstances
	}

	attachedInstances = append(attachedInstances, gpu.AttachedInstance{
		Id:                 attachedInstance.Id,
		Name:               attachedInstance.Name,
		ProfileAlias:       nil,
		UtilizationPercent: deviceGpuUtilizationRate,
		MemoryUsage: gpu.InstanceMemoryUsage{
			AllocatedMiB: deviceMemoryUsedMiB,
			TotalMiB:     deviceMemoryTotalMiB,
		},
		// hex supplies a pgpu's instance name, so the prefetched server map is
		// consulted only for the owning project the Grafana link needs. It is
		// absent on a pgpu-only node, where the prefetch is skipped entirely.
		Links: buildInstanceLinks(instanceLinkOpts{
			InstanceId: attachedInstance.Id,
			TenantId:   opts.Servers[attachedInstance.Id].TenantId,
			VmName:     attachedInstance.Name,
		}),
	})

	return &attachedInstances
}

// listVgpuAttachedInstances returns the attached instances for SR-IOV and
// MIG-backed vGPUs, sourced from one `nvidia-smi vgpu -q` snapshot already
// scoped to this device (opts.VgpuInstances). Unlike the old per-instance NVML
// calls, this snapshot is captured atomically by a single subprocess, so there
// is no "stale handle from a VM torn down mid-enumeration" case to guard
// against: whatever nvidia-smi saw at invocation time is what gets reported.
// A failed query still degrades to a partial or empty list instead of failing
// the whole node listing, so a single GPU being reset does not take down the
// listing for every other card.
func listVgpuAttachedInstances(opts listAttachedInstancesOpts) *[]gpu.AttachedInstance {
	deviceUUID, hexProfilesMap, serversById, enr := opts.DeviceUUID, opts.HexProfilesMap, opts.Servers, opts.Enrichment

	attachedInstances := []gpu.AttachedInstance{}
	if !opts.IsDeviceVisible {
		// The device was unavailable; the caller already flagged the card
		// degraded when it failed to resolve the device, so do not double-flag.
		return &attachedInstances
	}

	if !opts.VgpuInstancesAvailable {
		enr.degrade("nvidiasmi: failed to query vgpu instances for device %s; reporting no attached instances (degraded)", deviceUUID)
		return &attachedInstances
	}

	for _, instance := range opts.VgpuInstances {
		// hex reports both vGPU flavours' profile ids as vGPU Type IDs, which is
		// exactly what `nvidia-smi vgpu -q` gives per active instance - so the
		// lookup is direct for MIG-backed cards too. It used to need a
		// type -> GPU Instance Profile ID translation, because
		// gpu_vgpu_profile_list keyed migBacked entries by partition shape; that
		// list is per vGPU type as of cubecos #905.
		hexProfile := hexProfilesMap[instance.VgpuTypeId]
		profileAlias := hexProfile.Alias

		// The name and owning project are enrichment resolved from the prefetched
		// map. A missing entry just logs here (the instance is reported without a
		// name, and its link cannot pin the dashboard's tenant); a nil map means the
		// prefetch failed outright, which the caller already flagged as degraded on
		// the card itself.
		server, ok := serversById[instance.VmUUID]
		if !ok {
			log.Warnf("gpu: Openstack server %s not found in prefetched servers; reporting instance without name", instance.VmUUID)
		}

		attachedInstances = append(attachedInstances, gpu.AttachedInstance{
			Id:                 instance.VmUUID,
			Name:               server.Name,
			ProfileAlias:       profileAlias,
			UtilizationPercent: instance.GpuUtilizationPercent,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: instance.MemoryUsedMiB,
				TotalMiB:     instance.MemoryTotalMiB,
			},
			Links: buildInstanceLinks(instanceLinkOpts{InstanceId: instance.VmUUID, TenantId: server.TenantId, VmName: server.Name}),
		})
	}

	return &attachedInstances
}

// buildGpuCardLinksViaGrafana builds the history links reported inline with each
// GPU card. Both are pure string building from data already in hand -- the node
// name and the card's PCI address -- so they cost nothing per card and need no
// extra endpoint for the UI to call per row.
func buildGpuCardLinksViaGrafana(hostname, pciAddress string) gpu.GpuCardLinks {
	return gpu.GpuCardLinks{
		WorkloadHistory: grafana.DeviceGpuUtilizationHistoryLink(hostname, pciAddress),
		VramHistory:     grafana.DeviceGpuVramHistoryLink(hostname, pciAddress),
	}
}

// instanceLinkOpts is what the Grafana deep link needs about one attached
// instance. It is a struct rather than positional arguments so that adding a
// dashboard variable later does not ripple through every call site and test
// seam.
type instanceLinkOpts struct {
	InstanceId string
	TenantId   string
	VmName     string
}

// buildInstanceLinksViaOpenstack builds the two history links reported inline
// with each attached instance, one per panel of the instance dashboard's vGPU
// row. Both are pure string building from data already in hand, so they cost
// nothing per instance and need no extra endpoint for the UI to call per row --
// the same shape as the card-level pair.
//
// It deliberately carries no console link: creating a Nova
// console is a write that mints a short-lived token, and doing it for every
// instance on every GPU-list poll floods Nova with sessions that are almost
// always discarded. The console is minted on demand via the dedicated
// getGpuInstanceConsole endpoint instead.
func buildInstanceLinksViaOpenstack(opts instanceLinkOpts) gpu.InstanceLinks {
	// Both variables are required for the link to survive the dashboard's own
	// variable refresh. A pgpu-only node skips the Openstack prefetch, so the
	// tenant is unknown there; an instance missing from the prefetch has neither.
	// Reporting no link beats reporting one that lands on another VM, and a pgpu
	// has no vGPU series to chart in the first place.
	if opts.TenantId == "" || opts.VmName == "" {
		return gpu.InstanceLinks{}
	}

	workload := grafana.InstanceVgpuWorkloadHistoryLink(opts.InstanceId, opts.TenantId, opts.VmName)
	vram := grafana.InstanceVgpuVramHistoryLink(opts.InstanceId, opts.TenantId, opts.VmName)

	return gpu.InstanceLinks{WorkloadHistory: &workload, VramHistory: &vram}
}

// getGpuInstanceConsole mints a Nova console for a single attached instance on
// demand. Console tokens are short-lived, so they are created only when the user
// actually asks to open a console rather than during every GPU listing.
func (h *helper) getGpuInstanceConsole() (*gpu.InstanceConsole, error) {
	console, err := createConsole(h.instanceId)
	if err != nil {
		return nil, err
	}
	if console == nil {
		return nil, fmt.Errorf("gpu: no console returned for instance %s", h.instanceId)
	}

	return &gpu.InstanceConsole{Console: console.URL}, nil
}

func createConsoleViaOpenstack(vmId string) (*remoteconsoles.RemoteConsole, error) {
	// Nova only routes POST /servers/{id}/remote-consoles from microversion 2.6.
	// The shared compute client sends no microversion header, so Nova falls back
	// to 2.1 and answers 404. Copy the client and raise the microversion here
	// instead of mutating the shared one, which every other caller reads.
	compute := *openstack.GetGlobalHelper().Compute
	compute.Microversion = remoteConsoleMicroversion

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	result := remoteconsoles.Create(ctx, &compute, vmId, remoteconsoles.CreateOpts{
		Protocol: remoteconsoles.ConsoleProtocolVNC,
		Type:     remoteconsoles.ConsoleTypeNoVNC,
	})

	return result.Extract()
}

// getOpenstackServersViaHelper returns a map of server id to the enrichment an
// attached instance needs, in a single Openstack round trip.
func getOpenstackServersViaHelper() (map[string]openstackServer, error) {
	// AllTenants is required: the API authenticates as a single project (admin),
	// so a plain list returns only that project's servers and a vGPU VM owned by
	// any other project would come back nameless. The per-instance GetServer this
	// prefetch replaced was unaffected, because fetching one server by id is not
	// project-scoped for an admin token.
	serverList, err := openstack.GetGlobalHelper().ListServers(servers.ListOpts{AllTenants: true})
	if err != nil {
		return nil, err
	}

	// Not named `servers`: that is the gophercloud package used just above.
	byId := make(map[string]openstackServer, len(serverList))
	for _, server := range serverList {
		byId[server.ID] = openstackServer{Name: server.Name, TenantId: server.TenantID}
	}

	return byId, nil
}

func toProfileCollection(
	hexProfileCollection gpu.VgpuProfileCollectionFromHex,
	attachedInstances *[]gpu.AttachedInstance,
) gpu.GpuProfileCollection {
	collection := gpu.GpuProfileCollection{
		SriovVgpu:     []gpu.VgpuProfile{},
		MigBackedVgpu: []gpu.VgpuProfile{},
	}

	if hexProfileCollection.Sriov != nil {
		for _, profile := range *hexProfileCollection.Sriov {
			collection.SriovVgpu = append(collection.SriovVgpu, gpu.VgpuProfile{
				Id:         profile.Id,
				Name:       profile.Name,
				VramMiB:    uint64(profile.VramMiB),
				Count:      profile.Count,
				Remaining:  nil,
				AliasName:  profile.Alias,
				CountLimit: profile.VmCountLimit,
			})
		}
	}

	migProfileRemainingMap := createMigProfileRemainingMap(hexProfileCollection.MigBacked, attachedInstances)

	if hexProfileCollection.MigBacked != nil {
		for _, profile := range *hexProfileCollection.MigBacked {
			remaining := migProfileRemainingMap[profile.Id]

			collection.MigBackedVgpu = append(collection.MigBackedVgpu, gpu.VgpuProfile{
				Id:         profile.Id,
				Name:       profile.Name,
				VramMiB:    uint64(profile.VramMiB),
				Count:      profile.Count,
				Remaining:  &remaining,
				AliasName:  profile.Alias,
				CountLimit: profile.VmCountLimit,
			})
		}
	}

	return collection
}

// Returns a map with profile ID as key, and remaining count as value.
func createMigProfileRemainingMap(
	migProfiles *[]gpu.VgpuProfileFromHex,
	attachedInstances *[]gpu.AttachedInstance,
) map[uint32]int {
	// Key: profile ID. Value: remaining count.
	remainingMap := map[uint32]int{}

	if migProfiles == nil || attachedInstances == nil {
		return remainingMap
	}

	// Key: profile alias. Value: profile ID.
	profileIdMap := map[string]uint32{}

	for _, profile := range *migProfiles {
		if profile.Alias == nil || len(*profile.Alias) == 0 {
			continue
		}
		remainingMap[profile.Id] = profile.Count
		profileIdMap[*profile.Alias] = profile.Id
	}

	for _, instance := range *attachedInstances {
		if instance.ProfileAlias == nil || len(*instance.ProfileAlias) == 0 {
			continue
		}

		profileId, exists := profileIdMap[*instance.ProfileAlias]
		if exists {
			remainingMap[profileId] = max(remainingMap[profileId]-1, 0)
		}
	}

	return remainingMap
}

func (h *helper) updateNodeGpuCard() error {
	if nodes.IsLocal(h.node) {
		return h.updateLocalGpuCard()
	}
	return h.updateRemoteGpuCard()
}

func (h *helper) updateLocalGpuCard() error {
	card, err := getNodeGpuById(h.node, h.gpuId)
	if err != nil {
		return err
	}

	// vGPU profile/VRAM limits need the GPU's available profiles (hex) and its
	// total VRAM (nvidia-smi). Only gather them when profiles are actually supplied.
	var profilesMap map[uint32]gpu.VgpuProfileFromHex
	var totalVramMiB int
	if isVgpuType(h.gpuCardReq.ResourceType) && len(h.gpuCardReq.Profiles) > 0 {
		profilesMap, _, err = getNodeVgpuProfilesMap(card.Id)
		if err != nil {
			log.Errorf("gpu(%s): failed to get vgpu profiles for gpu %s: %v", h.reqId, h.gpuId, err)
			return err
		}

		totalVramMiB, err = getGpuTotalVramMiB(card.Id)
		if err != nil {
			log.Errorf("gpu(%s): failed to get total vram for gpu %s: %v", h.reqId, h.gpuId, err)
			return err
		}
	}

	if err := validateGpuCardUpdate(card, h.gpuCardReq, profilesMap, totalVramMiB); err != nil {
		return err
	}

	if err := upsertUpdatingGpuReq(h, h.gpuId); err != nil {
		return err
	}
	defer func() {
		if err := deleteUpdatingGpuReq(h, h.gpuId); err != nil {
			log.Errorf("gpu(%s): failed to clear updating record for gpu %s: %v", h.reqId, h.gpuId, err)
		}
	}()

	return updateNodeGpuCardViaHex(h.gpuId, h.gpuCardReq)
}

func (h *helper) updateRemoteGpuCard() error {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("gpu(%s): failed to get node %s: %v", h.reqId, h.node, err)
		return err
	}

	resp, err := h.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(h.gpuCardReq).
		Put(node.UpdateGpuCardUrl(h.gpuId))
	if err != nil {
		log.Errorf("gpu(%s): failed to update GPU card on remote node %s: %v", h.reqId, h.node, err)
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error response from node %s updating GPU card %s: %s", h.node, h.gpuId, string(resp.Body()))
	}

	return nil
}

func isGpuUpdatingViaMongo(h *helper, gpuId string) bool {
	count, err := h.mongo.GetCount(
		nodes.Db,
		nodes.ReqGpuCollection,
		bson.M{"hostname": h.node, "gpuId": gpuId},
	)
	if err != nil {
		log.Errorf("gpu(%s): failed to get updating record for gpu %s: %v", h.reqId, gpuId, err)
		return false
	}

	return count > 0
}

func upsertUpdatingGpuReqViaMongo(h *helper, gpuId string) error {
	record := gpu.GpuReqRecord{
		Hostname:     h.node,
		GpuId:        gpuId,
		ReqId:        h.reqId,
		ResourceType: h.gpuCardReq.ResourceType,
		Profiles:     h.gpuCardReq.Profiles,
	}

	return h.mongo.UpdateOne(
		nodes.Db,
		nodes.ReqGpuCollection,
		bson.M{"hostname": h.node, "gpuId": gpuId, "reqId": h.reqId},
		bson.M{"$set": record},
		options.Update().SetUpsert(true),
	)
}

func deleteUpdatingGpuReqViaMongo(h *helper, gpuId string) error {
	return h.mongo.DeleteOne(
		nodes.Db,
		nodes.ReqGpuCollection,
		bson.M{"hostname": h.node, "gpuId": gpuId, "reqId": h.reqId},
	)
}

// getGpuTotalVramMiB reads the physical GPU's total VRAM via nvidia-smi. Used
// as the budget for the vGPU VRAM-limit validation.
func getGpuTotalVramMiB(uuid string) (int, error) {
	devices, err := getNvidiaSmiDevices()
	if err != nil {
		return 0, fmt.Errorf("nvidiasmi: failed to query devices for gpu %s: %w", uuid, err)
	}

	device, ok := devices[uuid]
	if !ok {
		return 0, fmt.Errorf("nvidiasmi: device %s not reported by nvidia-smi", uuid)
	}

	return device.MemoryTotalMiB, nil
}
