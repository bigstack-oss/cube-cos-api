package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"math"
	"reflect"
	"slices"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/nvmlruntime"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Seams for unit tests: hex_sdk CLI, the NVML driver, and Openstack are
// unavailable there.
var (
	getNodeGpusMap              = cubecos.GetNodeGpusMap
	getNodeVgpuProfilesMap      = cubecos.GetNodeVgpuProfilesMap
	getNodePgpuAttachedInstance = cubecos.GetNodePgpuAttachedInstance
	deviceGetHandleByUUID       = nvml.DeviceGetHandleByUUID
	deviceGetCount              = nvml.DeviceGetCount
	deviceGetHandleByIndex      = nvml.DeviceGetHandleByIndex
	buildInstanceLinks          = buildInstanceLinksViaOpenstack
	getOpenstackServerNames     = getOpenstackServerNamesViaHelper
	createConsole               = createConsoleViaOpenstack
	vgpuInstanceId              = vgpuInstanceIdViaReflect
	isNvmlAvailable             = nvmlruntime.IsAvailable
	getNodeGpuById              = cubecos.GetNodeGpuById
	updateNodeGpuCardViaHex     = cubecos.UpdateNodeGpuCard
	isGpuUpdating               = isGpuUpdatingViaMongo
	upsertUpdatingGpuReq        = upsertUpdatingGpuReqViaMongo
	deleteUpdatingGpuReq        = deleteUpdatingGpuReqViaMongo
)

type listAttachedInstancesOpts struct {
	Device                   nvml.Device
	IsDeviceVisible          bool
	DeviceUUID               string
	DeviceMemoryUsedMiB      int
	DeviceMemoryTotalMiB     int
	DeviceGpuUtilizationRate uint32
	NodeName                 string
	HexGpu                   gpu.GpuFromHex
	HexProfilesMap           map[uint32]gpu.VgpuProfileFromHex
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

// enrichment accumulates the degraded state of a single GPU card while its NVML,
// hex and Openstack enrichment is gathered. degrade flags the card and logs why;
// coupling flag and log in one call keeps a "(degraded)" warning from ever
// drifting apart from the flag it is supposed to accompany. Gaps that are merely
// incomplete but not misleading (an instance reported without a name, a transient
// hex read skew) log directly via log.Warnf without flagging.
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

	gpuCards := []gpu.GpuCard{}

	// Iterate over hex GPUs instead of NVML devices: a GPU passed through to
	// a VM is invisible to NVML, but must still be reported.
	for _, pciAddress := range slices.Sorted(maps.Keys(hexGpusMap)) {
		gpuCard, err := h.buildLocalGpuCard(hexGpusMap[pciAddress], serverNames)
		if err != nil {
			return nil, err
		}

		gpuCards = append(gpuCards, gpuCard)
	}

	// Hex is the source of truth for the response, but a GPU visible to NVML yet
	// missing from hex points at a stale hex inventory (hot-add, cache lag): warn
	// so the mismatch is not entirely silent.
	h.warnGpusMissingFromHex(hexGpusMap)

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

func (h *helper) buildLocalGpuCard(hexGpu gpu.GpuFromHex, serverNames map[string]string) (gpu.GpuCard, error) {
	var memoryUsedMiB, memoryTotalMiB int
	var memoryUtilizationPercent, gpuUtilizationPercent uint32
	enr := &enrichment{}

	// Relies on hex reporting the GPU id as the NVML UUID.
	device, ret := deviceGetHandleByUUID(hexGpu.Id)
	isDeviceVisible := ret == nvml.SUCCESS

	// NVML runtime stats are enrichment on top of the hex inventory: when they
	// are unavailable the card is still reported, just without stats.
	switch {
	case isDeviceVisible:
		memoryInfo, ret := device.GetMemoryInfo()
		if ret == nvml.SUCCESS {
			memoryUsedMiB = bytesToMiB(memoryInfo.Used)
			memoryTotalMiB = bytesToMiB(memoryInfo.Total)
		} else {
			// Memory info is the card's capacity; losing it leaves the reported
			// capacity untrustworthy, so flag the card degraded.
			enr.degrade("nvml: failed to get memory info for device %s: %s; reporting card as degraded", hexGpu.Id, nvml.ErrorString(ret))
		}

		utilizationRates, ret := device.GetUtilizationRates()
		if ret == nvml.SUCCESS {
			memoryUtilizationPercent = utilizationRates.Memory
			gpuUtilizationPercent = utilizationRates.Gpu
		} else {
			// Utilization is runtime enrichment this card should have had; its
			// absence makes the card's stats untrustworthy, so flag it degraded.
			enr.degrade("nvml: failed to get utilization rates for device %s: %s; reporting card as degraded", hexGpu.Id, nvml.ErrorString(ret))
		}
	case ret == nvml.ERROR_NOT_FOUND && hexGpu.Type == gpu.ResourceTypePgpu:
		// Expected: a pgpu bound to vfio for passthrough (attached to a VM or
		// reserved for one) is invisible to NVML. Only a pgpu can disappear this
		// way, so any other type not being found is genuinely unexpected and
		// falls through to the warning below. Logged so an unexpected hex-id vs
		// NVML-UUID mismatch (which also surfaces as ERROR_NOT_FOUND) is not
		// entirely silent.
		log.Debugf("nvml: pgpu %s not visible to NVML (expected for vfio passthrough); reporting from hex without runtime stats", hexGpu.Id)
	default:
		// NVML was expected to see this device but could not provide a handle, so
		// the card is reported without runtime stats or attached instances and
		// its capacity is untrustworthy: flag it degraded. A node-wide NVML outage
		// (init failed) is the common cause, so distinguish it in the log.
		if !isNvmlAvailable() {
			enr.degrade("nvml: NVML is not initialized on this node; gpu %s reported from hex without runtime stats or attachments (capacity is degraded)", hexGpu.Id)
		} else {
			enr.degrade("nvml: failed to get device handle for gpu %s: %s; reporting card as degraded", hexGpu.Id, nvml.ErrorString(ret))
		}
	}

	hexProfilesMap := map[uint32]gpu.VgpuProfileFromHex{}
	hexProfileCollection := gpu.VgpuProfileCollectionFromHex{}

	if isVgpu(hexGpu) {
		var profErr error
		hexProfilesMap, hexProfileCollection, profErr = getNodeVgpuProfilesMap(hexGpu.PciAddress)
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
	if serverNames == nil {
		enr.degrade("gpu: Openstack server prefetch failed on node %s; gpu %s reported without instance names (degraded)", h.node, hexGpu.Id)
	}

	attachedInstances, err := listAttachedInstances(listAttachedInstancesOpts{
		Device:                   device,
		IsDeviceVisible:          isDeviceVisible,
		DeviceUUID:               hexGpu.Id,
		DeviceMemoryUsedMiB:      memoryUsedMiB,
		DeviceMemoryTotalMiB:     memoryTotalMiB,
		DeviceGpuUtilizationRate: gpuUtilizationPercent,
		NodeName:                 h.node,
		HexGpu:                   hexGpu,
		HexProfilesMap:           hexProfilesMap,
		ServerNames:              serverNames,
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

// warnGpusMissingFromHex enumerates NVML devices and warns for any that hex did
// not report. Hex drives the response, so an NVML-visible GPU absent from hex
// would otherwise silently disappear; the old code hard-errored on this. Runs
// only when NVML is available (a vfio-passthrough GPU is invisible to NVML and
// legitimately absent from enumeration, so it never triggers a warning).
func (h *helper) warnGpusMissingFromHex(hexGpusMap map[string]gpu.GpuFromHex) {
	if !isNvmlAvailable() {
		return
	}

	count, ret := deviceGetCount()
	if ret != nvml.SUCCESS {
		log.Warnf("nvml: failed to get device count for hex reconciliation on node %s: %s", h.node, nvml.ErrorString(ret))
		return
	}

	hexUUIDs := map[string]struct{}{}
	for _, hexGpu := range hexGpusMap {
		hexUUIDs[hexGpu.Id] = struct{}{}
	}

	for i := range count {
		device, ret := deviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get device handle at index %d for hex reconciliation on node %s: %s", i, h.node, nvml.ErrorString(ret))
			continue
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get UUID for device at index %d for hex reconciliation on node %s: %s", i, h.node, nvml.ErrorString(ret))
			continue
		}

		if _, ok := hexUUIDs[uuid]; !ok {
			log.Warnf("gpu: NVML reports device %s that is absent from the hex inventory on node %s; hex may be stale (hot-add, cache lag)", uuid, h.node)
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

func bytesToMiB(bytes uint64) int {
	return int(bytes / 1024 / 1024)
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
// capabilities. profilesMap (keyed by profile Id) and totalVramMiB (from NVML)
// are only consulted for vGPU types with profiles. Returns a sentinel-wrapped
// error the handler maps to a status.
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

	// The profile-count limit only applies to SR-IOV vGPU; MIG-backed vGPU has no such limit.
	if req.ResourceType == gpu.ResourceTypeSriovVgpu &&
		card.SriovVgpuProfileCountLimit != nil && len(distinctIds) > *card.SriovVgpuProfileCountLimit {
		return fmt.Errorf("gpu %s: %d distinct profiles exceed profile count limit %d: %w",
			card.Id, len(distinctIds), *card.SriovVgpuProfileCountLimit, gpu.ErrExceedProfileCountLimit)
	}

	totalVramReqMiB := 0
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
// opts.Enrichment when NVML/hex enrichment that should have populated them could
// not be obtained. An error is reserved for integrity failures (an unhandled
// type) that must fail the whole listing.
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
// MIG-backed vGPUs. The instance list is derived entirely from NVML (enrichment):
// any NVML failure degrades to a partial or empty list instead of failing the
// whole node listing, so a single GPU being reset does not take down the listing
// for every other card.
func listVgpuAttachedInstances(opts listAttachedInstancesOpts) *[]gpu.AttachedInstance {
	device, deviceUUID, hexProfilesMap, serverNames, enr := opts.Device, opts.DeviceUUID, opts.HexProfilesMap, opts.ServerNames, opts.Enrichment

	attachedInstances := []gpu.AttachedInstance{}
	if !opts.IsDeviceVisible {
		// The device handle was unavailable; the caller already flagged the card
		// degraded when it failed to resolve the handle, so do not double-flag.
		return &attachedInstances
	}

	vgpuInstances, ret := device.GetActiveVgpus()
	if ret != nvml.SUCCESS {
		// The active vGPU list is enrichment (e.g. ERROR_GPU_IS_LOST during a
		// reset): degrade to no attached instances rather than failing the whole
		// node listing.
		enr.degrade("nvml: failed to get active vgpus for device %s: %s; reporting no attached instances (degraded)", deviceUUID, nvml.ErrorString(ret))
		return &attachedInstances
	}

	// No active vGPU: skip the utilization query entirely (it can fail on a
	// device with no samples yet, and its result would be unused anyway).
	if len(vgpuInstances) == 0 {
		return &attachedInstances
	}

	vgpuInstanceUtilizationMap := buildVgpuInstanceUtilizationMap(device, deviceUUID, enr)

	for i, vgpuInstance := range vgpuInstances {
		// The per-instance NVML calls below are enrichment on top of the active
		// vGPU list: a VM torn down between GetActiveVgpus and these calls leaves
		// a stale handle that returns ERROR_NOT_FOUND. Skip just that instance
		// (flagging the card degraded so the missing instance is not read as a
		// real detach) instead of failing the whole node listing.
		vmId, _, ret := vgpuInstance.GetVmID()
		if ret != nvml.SUCCESS {
			enr.degrade("nvml: failed to get VM ID for vgpu instance at index %d: %s; skipping instance (degraded)", i, nvml.ErrorString(ret))
			continue
		}

		vgpuType, ret := vgpuInstance.GetType()
		if ret != nvml.SUCCESS {
			enr.degrade("nvml: failed to get type for vgpu instance %s: %s; skipping instance (degraded)", vmId, nvml.ErrorString(ret))
			continue
		}

		profileId, ret := vgpuType.GetGpuInstanceProfileId()
		if ret != nvml.SUCCESS {
			enr.degrade("nvml: failed to get profile ID for vgpu instance %s: %s; skipping instance (degraded)", vmId, nvml.ErrorString(ret))
			continue
		}

		frameBufferBytes, ret := vgpuType.GetFramebufferSize()
		if ret != nvml.SUCCESS {
			enr.degrade("nvml: failed to get frame buffer size for profile %d: %s; skipping instance %s (degraded)", profileId, nvml.ErrorString(ret), vmId)
			continue
		}

		fbUsage, ret := vgpuInstance.GetFbUsage()
		if ret != nvml.SUCCESS {
			enr.degrade("nvml: failed to get fb usage for vgpu instance %s: %s; skipping instance (degraded)", vmId, nvml.ErrorString(ret))
			continue
		}

		hexProfile := hexProfilesMap[profileId]
		profileAlias := hexProfile.Alias

		// The server name is enrichment resolved from the prefetched map. A missing
		// name just logs here (the instance is reported without one); a nil map
		// means the prefetch failed outright, which the caller already flagged as
		// degraded on the card itself.
		serverName, ok := serverNames[vmId]
		if !ok {
			log.Warnf("gpu: Openstack server %s not found in prefetched servers; reporting instance without name", vmId)
		}

		utilizationPercent := vgpuInstanceUtilizationMap[vgpuInstanceId(vgpuInstance)]

		attachedInstances = append(attachedInstances, gpu.AttachedInstance{
			Id:                 vmId,
			Name:               serverName,
			ProfileAlias:       profileAlias,
			UtilizationPercent: utilizationPercent,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: bytesToMiB(fbUsage),
				TotalMiB:     bytesToMiB(frameBufferBytes),
			},
			Links: buildInstanceLinks(vmId),
		})
	}

	return &attachedInstances
}

// buildVgpuInstanceUtilizationMap maps vGPU instance handle to SM utilization.
// A failed query would leave every attached instance reporting 0% utilization,
// which reads as idle rather than unknown, so any non-SUCCESS degrades the card
// (never fails the listing) and returns an empty map.
func buildVgpuInstanceUtilizationMap(device nvml.Device, deviceUUID string, enr *enrichment) map[uint32]uint32 {
	utilizationMap := map[uint32]uint32{}
	valueType, samples, ret := device.GetVgpuUtilization(0)

	if ret != nvml.SUCCESS {
		enr.degrade("nvml: failed to get vgpu utilization for device %s: %s; reporting empty utilization (degraded)", deviceUUID, nvml.ErrorString(ret))
		return utilizationMap
	}

	for _, sample := range samples {
		switch valueType {
		case nvml.VALUE_TYPE_UNSIGNED_INT:
			utilizationMap[sample.VgpuInstance] = binary.LittleEndian.Uint32(sample.SmUtil[:4])
		case nvml.VALUE_TYPE_DOUBLE:
			// SmUtil holds the raw bytes of an IEEE-754 double; reinterpret it
			// rather than reading the low bytes as an integer.
			doubleBits := binary.LittleEndian.Uint64(sample.SmUtil[:])
			utilizationMap[sample.VgpuInstance] = uint32(math.Round(math.Float64frombits(doubleBits)))
		}
	}

	return utilizationMap
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

// nvml.VgpuInstance does not expose its raw handle; the driver's concrete
// type is an integer handle, recovered here via reflection.
func vgpuInstanceIdViaReflect(vgpuInstance nvml.VgpuInstance) uint32 {
	return uint32(reflect.ValueOf(vgpuInstance).Uint())
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
				VramMiB:    profile.VramMiB,
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
				VramMiB:    profile.VramMiB,
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
	// total VRAM (NVML). Only gather them when profiles are actually supplied.
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

// getGpuTotalVramMiB reads the physical GPU's total VRAM via NVML. Used as the
// budget for the vGPU VRAM-limit validation.
func getGpuTotalVramMiB(uuid string) (int, error) {
	device, ret := deviceGetHandleByUUID(uuid)
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("nvml: failed to get device handle for gpu %s: %s", uuid, nvml.ErrorString(ret))
	}

	memoryInfo, ret := device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("nvml: failed to get memory info for gpu %s: %s", uuid, nvml.ErrorString(ret))
	}

	return bytesToMiB(memoryInfo.Total), nil
}
