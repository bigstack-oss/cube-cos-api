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

// Seams for unit tests: hex_sdk CLI, nvidia-smi, and Openstack are
// unavailable there.
var (
	getNodeGpusMap                 = cubecos.GetNodeGpusMap
	getNodeVgpuProfilesMap         = cubecos.GetNodeVgpuProfilesMap
	getNodePgpuAttachedInstance    = cubecos.GetNodePgpuAttachedInstance
	getNvidiaSmiDevices            = cubecos.GetNvidiaSmiDevices
	getNvidiaSmiVgpuInstances      = cubecos.GetNvidiaSmiVgpuInstances
	getNvidiaSmiVgpuTypeProfileIds = cubecos.GetNvidiaSmiVgpuTypeGpuInstanceProfileIds
	buildInstanceLinks             = buildInstanceLinksViaOpenstack
	getOpenstackServerNames        = getOpenstackServerNamesViaHelper
	createConsole                  = createConsoleViaOpenstack
	getNodeGpuById                 = cubecos.GetNodeGpuById
	updateNodeGpuCardViaHex        = cubecos.UpdateNodeGpuCard
	isGpuUpdating                  = isGpuUpdatingViaMongo
	upsertUpdatingGpuReq           = upsertUpdatingGpuReqViaMongo
	deleteUpdatingGpuReq           = deleteUpdatingGpuReqViaMongo
)

// buildLocalGpuCardOpts carries data fetched once per listLocalGpuCards
// request (Openstack server names, nvidia-smi's device and vGPU-instance
// snapshots) into each card's build. nvidia-smi is a subprocess spawn, not an
// in-process library call, so it is invoked once per request here rather than
// once per GPU on the node.
type buildLocalGpuCardOpts struct {
	ServerNames map[string]string
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
	IsDeviceVisible          bool
	DeviceUUID               string
	DeviceMemoryUsedMiB      int
	DeviceMemoryTotalMiB     int
	DeviceGpuUtilizationRate uint32
	NodeName                 string
	HexGpu                   gpu.GpuFromHex
	HexProfilesMap           map[uint32]gpu.VgpuProfileFromHex
	// VgpuInstances/VgpuInstancesAvailable: see buildLocalGpuCardOpts.
	VgpuInstances          []cubecos.NvidiaSmiVgpuInstance
	VgpuInstancesAvailable bool
	// MigTypeProfileIds maps vGPU Type ID -> GPU Instance Profile ID for a
	// MIG-backed GPU's supported types (nvidia-smi vgpu -q reports the former
	// per instance; hex's MIG-backed profile ids are the latter). Unused for
	// SR-IOV, whose hex profile ids are the vGPU Type ID directly.
	MigTypeProfileIds map[uint32]uint32
	// ServerNames maps Openstack server id to name, prefetched once per request
	// so vGPU instance names do not cost a GetServer round trip each. A nil map
	// means the prefetch failed outright, so a name missing for an attached
	// instance is a real enrichment loss (degrade); a non-nil map that simply
	// lacks an id means that one VM is absent (soft, reported without a name).
	ServerNames map[string]string
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
	serverNames := h.resolveServerNames(hexGpusMap)

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
			ServerNames:            serverNames,
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

// resolveServerNames prefetches Openstack server names (id -> name) when the
// node has any vGPU, whose attached-instance names come from Openstack. On a
// lookup failure it returns nil (instances reported without a name) rather than
// failing the listing; the nil map lets callers mark affected cards degraded.
func (h *helper) resolveServerNames(hexGpusMap map[string]gpu.GpuFromHex) map[string]string {
	needsServerNames := false
	for _, hexGpu := range hexGpusMap {
		if isVgpu(hexGpu) {
			needsServerNames = true
			break
		}
	}
	if !needsServerNames {
		return map[string]string{}
	}

	serverNames, err := getOpenstackServerNames()
	if err != nil {
		log.Warnf("gpu(%s): failed to list Openstack servers for name resolution on node %s: %v; reporting instances without names", h.reqId, h.node, err)
		return nil
	}

	return serverNames
}

// resolveVgpuInstances fetches active vGPU instances once for the whole node
// when hex reports at least one vGPU-type GPU, skipping the nvidia-smi call
// entirely on pgpu-only nodes (mirrors resolveServerNames's lazy Openstack
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
	var memoryUsedMiB, memoryTotalMiB int
	var memoryUtilizationPercent, gpuUtilizationPercent uint32
	enr := &enrichment{}

	// Relies on hex reporting the GPU id as nvidia-smi's own UUID.
	device, isDeviceVisible := opts.NvidiaSmiDevices[hexGpu.Id]

	// nvidia-smi runtime stats are enrichment on top of the hex inventory:
	// when they are unavailable the card is still reported, just without
	// stats.
	switch {
	case isDeviceVisible:
		memoryUsedMiB = device.MemoryUsedMiB
		memoryTotalMiB = device.MemoryTotalMiB
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

	if isVgpu(hexGpu) {
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

	// A nil server-name map means the getOpenstackServerNames call failed, so
	// attached-instance name enrichment is unavailable for the whole node: flag
	// the card degraded.
	if opts.ServerNames == nil {
		enr.degrade("gpu: Openstack server prefetch failed on node %s; gpu %s reported without instance names (degraded)", h.node, hexGpu.Id)
	}

	// MIG-backed hex profile ids are the GPU Instance Profile ID, which
	// nvidia-smi only reports per vGPU *type* (vgpu -s -v), not per active
	// instance (vgpu -q reports the vGPU Type ID instead): resolve the
	// type -> profile-id mapping once here, only when actually needed.
	migTypeProfileIds := map[uint32]uint32{}
	if hexGpu.Type == gpu.ResourceTypeMigBackedVgpu && len(opts.VgpuInstances) > 0 {
		var err error
		migTypeProfileIds, err = getNvidiaSmiVgpuTypeProfileIds(hexGpu.PciAddress)
		if err != nil {
			enr.degrade("nvidiasmi: failed to get vgpu type profile ids for gpu %s: %v; mig-backed instance aliases may be missing (degraded)", hexGpu.Id, err)
		}
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
		MigTypeProfileIds:        migTypeProfileIds,
		ServerNames:              opts.ServerNames,
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
		Links: buildInstanceLinks(attachedInstance.Id),
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
	deviceUUID, hexProfilesMap, serverNames, enr := opts.DeviceUUID, opts.HexProfilesMap, opts.ServerNames, opts.Enrichment

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
		profileId := instance.VgpuTypeId

		if opts.HexGpu.Type == gpu.ResourceTypeMigBackedVgpu {
			giProfileId, ok := opts.MigTypeProfileIds[instance.VgpuTypeId]
			if !ok {
				enr.degrade("nvidiasmi: no GPU Instance Profile ID found for vgpu type %d on device %s; skipping instance %s (degraded)", instance.VgpuTypeId, deviceUUID, instance.VmUUID)
				continue
			}
			profileId = giProfileId
		}

		hexProfile := hexProfilesMap[profileId]
		profileAlias := hexProfile.Alias

		// The server name is enrichment resolved from the prefetched map. A missing
		// name just logs here (the instance is reported without one); a nil map
		// means the prefetch failed outright, which the caller already flagged as
		// degraded on the card itself.
		serverName, ok := serverNames[instance.VmUUID]
		if !ok {
			log.Warnf("gpu: Openstack server %s not found in prefetched servers; reporting instance without name", instance.VmUUID)
		}

		attachedInstances = append(attachedInstances, gpu.AttachedInstance{
			Id:                 instance.VmUUID,
			Name:               serverName,
			ProfileAlias:       profileAlias,
			UtilizationPercent: instance.GpuUtilizationPercent,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: instance.MemoryUsedMiB,
				TotalMiB:     instance.MemoryTotalMiB,
			},
			Links: buildInstanceLinks(instance.VmUUID),
		})
	}

	return &attachedInstances
}

// buildInstanceLinksViaOpenstack builds the links reported inline with each
// attached instance. It deliberately carries no console link: creating a Nova
// console is a write that mints a short-lived token, and doing it for every
// instance on every GPU-list poll floods Nova with sessions that are almost
// always discarded. The console is minted on demand via the dedicated
// getGpuInstanceConsole endpoint instead.
func buildInstanceLinksViaOpenstack(vmId string) gpu.InstanceLinks {
	return gpu.InstanceLinks{
		Grafana: grafana.InstanceDashboardLink(vmId),
	}
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
	openstackHelper := openstack.GetGlobalHelper()

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	result := remoteconsoles.Create(ctx, openstackHelper.Compute, vmId, remoteconsoles.CreateOpts{
		Protocol: remoteconsoles.ConsoleProtocolVNC,
		Type:     remoteconsoles.ConsoleTypeNoVNC,
	})

	return result.Extract()
}

// getOpenstackServerNamesViaHelper returns a map of server id to name in a single
// Openstack round trip, used to resolve vGPU attached-instance names in bulk.
func getOpenstackServerNamesViaHelper() (map[string]string, error) {
	serverList, err := openstack.GetGlobalHelper().ListServers(servers.ListOpts{})
	if err != nil {
		return nil, err
	}

	serverNames := make(map[string]string, len(serverList))
	for _, server := range serverList {
		serverNames[server.ID] = server.Name
	}

	return serverNames, nil
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
