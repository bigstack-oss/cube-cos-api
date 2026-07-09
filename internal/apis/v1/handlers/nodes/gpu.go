package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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
	buildInstanceLinks          = buildInstanceLinksViaOpenstack
	getOpenstackServer          = getOpenstackServerViaHelper
	createConsole               = createConsoleViaOpenstack
	vgpuInstanceId              = vgpuInstanceIdViaReflect
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

	gpuCards := []gpu.GpuCard{}

	// Iterate over hex GPUs instead of NVML devices: a GPU passed through to
	// a VM is invisible to NVML, but must still be reported.
	for _, pciAddress := range slices.Sorted(maps.Keys(hexGpusMap)) {
		gpuCard, err := h.buildLocalGpuCard(hexGpusMap[pciAddress])
		if err != nil {
			return nil, err
		}

		gpuCards = append(gpuCards, gpuCard)
	}

	return gpuCards, nil
}

func (h *helper) buildLocalGpuCard(hexGpu gpu.GpuFromHex) (gpu.GpuCard, error) {
	var memoryUsedMiB, memoryTotalMiB int
	var memoryUtilizationPercent, gpuUtilizationPercent uint32

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
		// falls through to the warning below.
	default:
		log.Warnf("nvml: failed to get device handle for gpu %s: %s", hexGpu.Id, nvml.ErrorString(ret))
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
	}

	return gpuCard, nil
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
		return nil, fmt.Errorf("gpu: hex gpu %s allocation.current is not 0, but its attached instance is null on node %s", hexGpu.PciAddress, nodeName)
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
	device, deviceUUID, hexProfilesMap := opts.Device, opts.DeviceUUID, opts.HexProfilesMap

	attachedInstances := []gpu.AttachedInstance{}
	if !opts.IsDeviceVisible {
		return &attachedInstances, nil
	}

	vgpuInstances, ret := device.GetActiveVgpus()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml: failed to get active vgpus for device %s: %s", deviceUUID, nvml.ErrorString(ret))
	}

	vgpuInstanceUtilizationMap, err := buildVgpuInstanceUtilizationMap(device, deviceUUID)
	if err != nil {
		return nil, err
	}

	for i, vgpuInstance := range vgpuInstances {
		vmId, _, ret := vgpuInstance.GetVmID()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("nvml: failed to get VM ID for vgpu instance at index %d: %s", i, nvml.ErrorString(ret))
		}

		vgpuType, ret := vgpuInstance.GetType()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("nvml: failed to get type for vgpu instance %s: %s", vmId, nvml.ErrorString(ret))
		}

		profileId, ret := vgpuType.GetGpuInstanceProfileId()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("nvml: failed to get profile ID for vgpu instance %s: %s", vmId, nvml.ErrorString(ret))
		}

		frameBufferBytes, ret := vgpuType.GetFramebufferSize()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("nvml: failed to get frame buffer size for profile %d: %s", profileId, nvml.ErrorString(ret))
		}

		fbUsage, ret := vgpuInstance.GetFbUsage()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("nvml: failed to get fb usage for vgpu instance %s: %s", vmId, nvml.ErrorString(ret))
		}

		hexProfile := hexProfilesMap[profileId]
		profileAlias := hexProfile.Alias

		// The server name is enrichment: the lookup can fail transiently
		// (e.g. the VM is being torn down), so degrade instead of failing
		// the whole listing.
		serverName := ""
		server, err := getOpenstackServer(vmId)
		switch {
		case err != nil:
			log.Warnf("gpu: failed to get Openstack server %s: %v", vmId, err)
		case server == nil:
			log.Warnf("gpu: Openstack server %s not found", vmId)
		default:
			serverName = server.Name
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

	// NVML returns ERROR_NOT_FOUND when no utilization samples exist (e.g.
	// no vGPU has been active since the last query); that is not a failure.
	if ret == nvml.ERROR_NOT_FOUND {
		return utilizationMap, nil
	}

	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml: failed to get vgpu utilization for device %s: %s", deviceUUID, nvml.ErrorString(ret))
	}

	for _, sample := range samples {
		switch valueType {
		case nvml.VALUE_TYPE_UNSIGNED_INT:
			utilizationMap[sample.VgpuInstance] = binary.LittleEndian.Uint32(sample.SmUtil[:4])
		case nvml.VALUE_TYPE_DOUBLE:
			utilizationMap[sample.VgpuInstance] = binary.LittleEndian.Uint32(sample.SmUtil[:])
		}
	}

	return utilizationMap, nil
}

func buildInstanceLinksViaOpenstack(vmId string) gpu.InstanceLinks {
	links := gpu.InstanceLinks{
		Grafana: grafana.InstanceDashboardLink(vmId),
	}

	// Nova refuses to create a console for a non-ACTIVE instance (409): the
	// console link is enrichment, so keep the remaining links instead of
	// failing the whole listing.
	console, err := createConsole(vmId)
	if err != nil {
		log.Warnf("openstack: failed to create console link for instance %s: %v", vmId, err)
		return links
	}
	if console == nil {
		log.Warnf("openstack: no console returned for instance %s", vmId)
		return links
	}

	links.Console = console.URL
	return links
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

func getOpenstackServerViaHelper(vmId string) (*servers.Server, error) {
	return openstack.GetGlobalHelper().GetServer(vmId)
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
