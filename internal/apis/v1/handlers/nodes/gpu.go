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
	// so vGPU instance names do not cost a GetServer round trip each.
	ServerNames map[string]string
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
	// one GetServer per attached instance.
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
// node has any vGPU, whose attached-instance names come from Openstack. A lookup
// failure degrades to an empty map (instances reported without a name) rather
// than failing the listing.
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
		return map[string]string{}
	}

	return serverNames
}

func (h *helper) buildLocalGpuCard(hexGpu gpu.GpuFromHex, serverNames map[string]string) (gpu.GpuCard, error) {
	var memoryUsedMiB, memoryTotalMiB int
	var memoryUtilizationPercent, gpuUtilizationPercent uint32
	var degraded bool

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
			log.Warnf("nvml: failed to get memory info for device %s: %s", hexGpu.Id, nvml.ErrorString(ret))
		}

		utilizationRates, ret := device.GetUtilizationRates()
		if ret == nvml.SUCCESS {
			memoryUtilizationPercent = utilizationRates.Memory
			gpuUtilizationPercent = utilizationRates.Gpu
		} else {
			log.Warnf("nvml: failed to get utilization rates for device %s: %s", hexGpu.Id, nvml.ErrorString(ret))
		}
		//TODO: We should add more check.
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
		degraded = true
		if !isNvmlAvailable() {
			log.Warnf("nvml: NVML is not initialized on this node; gpu %s reported from hex without runtime stats or attachments (capacity is degraded)", hexGpu.Id)
		} else {
			log.Warnf("nvml: failed to get device handle for gpu %s: %s; reporting card as degraded", hexGpu.Id, nvml.ErrorString(ret))
		}
	}

	hexProfilesMap := map[uint32]gpu.VgpuProfileFromHex{}
	hexProfileCollection := gpu.VgpuProfileCollectionFromHex{}

	if isVgpu(hexGpu) {
		hexProfilesMap, hexProfileCollection = getNodeVgpuProfilesMap(hexGpu.PciAddress)
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
			IsProcessing: false,
		},
		AllocationSummary:          hexGpu.Allocation,
		SriovVgpuProfileCountLimit: hexGpu.SriovVgpuProfileCountLimit,
		Profiles:                   profileCollection,
		AttachedInstances:          attachedInstances,
		Degraded:                   degraded,
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
	return hexGpu.Type == gpu.ResourceTypeSriovVgpu || hexGpu.Type == gpu.ResourceTypeMigBackedVgpu
}

func listAttachedInstances(opts listAttachedInstancesOpts) (*[]gpu.AttachedInstance, error) {
	hexGpu := opts.HexGpu

	switch opts.HexGpu.Type {
	case gpu.ResourceTypeUnset:
		return nil, nil
	case gpu.ResourceTypePgpu:
		return listPgpuAttachedInstances(opts)
	case gpu.ResourceTypeSriovVgpu, gpu.ResourceTypeMigBackedVgpu:
		return listVgpuAttachedInstances(opts)
	default:
		return nil, fmt.Errorf("gpu: unhandled gpu type %s when listing attached instances for gpu %s", hexGpu.Type, hexGpu.Id)
	}
}

func listPgpuAttachedInstances(opts listAttachedInstancesOpts) (*[]gpu.AttachedInstance, error) {
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
		return &attachedInstances, nil
	}

	attachedInstance, err := getNodePgpuAttachedInstance(hexGpu.PciAddress)
	if err != nil {
		return nil, err
	}

	if attachedInstance == nil {
		// The allocation count and the attached-instance lookup are two separate
		// hex reads: during an attach/detach they can be observed out of sync
		// (allocation already 1 while the instance is not yet reported). Degrade
		// to no attached instance instead of failing the whole node listing.
		log.Warnf("gpu: hex reports allocation for pgpu %s on node %s but no attached instance yet (transient read skew); reporting no attached instance", hexGpu.PciAddress, nodeName)
		return &attachedInstances, nil
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

	return &attachedInstances, nil
}

// Returns the attached instances for SR-IOV and MIG-backed vGPUs.
func listVgpuAttachedInstances(opts listAttachedInstancesOpts) (*[]gpu.AttachedInstance, error) {
	device, deviceUUID, hexProfilesMap, serverNames := opts.Device, opts.DeviceUUID, opts.HexProfilesMap, opts.ServerNames

	attachedInstances := []gpu.AttachedInstance{}
	if !opts.IsDeviceVisible {
		return &attachedInstances, nil
	}

	vgpuInstances, ret := device.GetActiveVgpus()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml: failed to get active vgpus for device %s: %s", deviceUUID, nvml.ErrorString(ret))
	}

	// No active vGPU: skip the utilization query entirely (it can fail on a
	// device with no samples yet, and its result would be unused anyway).
	if len(vgpuInstances) == 0 {
		return &attachedInstances, nil
	}

	vgpuInstanceUtilizationMap, err := buildVgpuInstanceUtilizationMap(device, deviceUUID)
	if err != nil {
		return nil, err
	}

	for i, vgpuInstance := range vgpuInstances {
		// The per-instance NVML calls below are enrichment on top of the active
		// vGPU list: a VM torn down between GetActiveVgpus and these calls leaves
		// a stale handle that returns ERROR_NOT_FOUND. Skip just that instance
		// instead of failing the whole node listing.
		vmId, _, ret := vgpuInstance.GetVmID()
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get VM ID for vgpu instance at index %d: %s; skipping instance", i, nvml.ErrorString(ret))
			continue
		}

		vgpuType, ret := vgpuInstance.GetType()
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get type for vgpu instance %s: %s; skipping instance", vmId, nvml.ErrorString(ret))
			continue
		}

		profileId, ret := vgpuType.GetGpuInstanceProfileId()
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get profile ID for vgpu instance %s: %s; skipping instance", vmId, nvml.ErrorString(ret))
			continue
		}

		frameBufferBytes, ret := vgpuType.GetFramebufferSize()
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get frame buffer size for profile %d: %s; skipping instance %s", profileId, nvml.ErrorString(ret), vmId)
			continue
		}

		fbUsage, ret := vgpuInstance.GetFbUsage()
		if ret != nvml.SUCCESS {
			log.Warnf("nvml: failed to get fb usage for vgpu instance %s: %s; skipping instance", vmId, nvml.ErrorString(ret))
			continue
		}

		hexProfile := hexProfilesMap[profileId]
		profileAlias := hexProfile.Alias

		// The server name is enrichment resolved from the prefetched map; a VM
		// being torn down (or briefly missing from the listing) simply has no
		// name rather than failing the whole listing.
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

	return &attachedInstances, nil
}

func buildVgpuInstanceUtilizationMap(device nvml.Device, deviceUUID string) (map[uint32]uint32, error) {
	utilizationMap := map[uint32]uint32{}
	valueType, samples, ret := device.GetVgpuUtilization(0)

	// Utilization is enrichment: NVML returns ERROR_NOT_FOUND when no samples
	// exist yet, ERROR_NOT_SUPPORTED on some hardware, and ERROR_GPU_IS_LOST
	// during a reset. None of those should fail the node listing, so any
	// non-SUCCESS degrades to an empty map.
	if ret != nvml.SUCCESS {
		log.Warnf("nvml: failed to get vgpu utilization for device %s: %s; reporting empty utilization", deviceUUID, nvml.ErrorString(ret))
		return utilizationMap, nil
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

	return utilizationMap, nil
}

// buildInstanceLinksViaOpenstack builds the links reported inline with each
// attached instance. The console link is intentionally NOT minted here: creating
// a Nova console is a write that mints a short-lived token, and doing it for
// every instance on every GPU-list poll floods Nova with sessions that are
// almost always discarded. The console is minted on demand via the dedicated
// getGpuInstanceConsole endpoint instead, so Console stays empty in list output.
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
