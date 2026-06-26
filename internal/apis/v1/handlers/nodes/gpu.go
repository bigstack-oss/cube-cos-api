package nodes

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
	log "go-micro.dev/v5/logger"
)

type listAttachedInstancesOpts struct {
	Device                   nvml.Device
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
	hexGpusMap, err := cubecos.GetNodeGpusMap(h.node)
	if err != nil {
		return nil, err
	}

	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		errorString := nvml.ErrorString(ret)
		log.Errorf("nvml: failed to get device count: %s", errorString)
		return nil, errors.New(errorString)
	}

	gpuCards := []gpu.GpuCard{}

	for i := range deviceCount {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			errorString := nvml.ErrorString(ret)
			log.Errorf("nvml: failed to get device handle at index %d: %s", i, errorString)
			continue
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			errorString := nvml.ErrorString(ret)
			log.Errorf("nvml: failed to get UUID for device at index %d: %s", i, errorString)
			continue
		}

		pciInfo, ret := device.GetPciInfo()
		if ret != nvml.SUCCESS {
			errorString := nvml.ErrorString(ret)
			log.Errorf("nvml: failed to get PCI info for device %s: %s", uuid, errorString)
			continue
		}

		memoryInfo, ret := device.GetMemoryInfo()
		var memoryUsedMiB, memoryTotalMiB int
		if ret == nvml.SUCCESS {
			memoryUsedMiB = bytesToMiB(memoryInfo.Used)
			memoryTotalMiB = bytesToMiB(memoryInfo.Total)
		} else {
			// Failed to get memory info is non-fatal.
			log.Errorf("nvml: failed to get memory info for device %s: %s", uuid, nvml.ErrorString(ret))
		}

		utilizationRates, ret := device.GetUtilizationRates()
		var memoryUtilizationPercent, gpuUtilizationPercent uint32
		if ret == nvml.SUCCESS {
			memoryUtilizationPercent = utilizationRates.Memory
			gpuUtilizationPercent = utilizationRates.Gpu
		} else {
			// Failed to get utilization rates is non-fatal.
			log.Errorf("nvml: failed to get utilization rates for device %s: %s", uuid, nvml.ErrorString(ret))
		}

		pciAddress := extractPciAddress(pciInfo)
		hexGpu := hexGpusMap[pciAddress]

		if len(hexGpu.Id) == 0 {
			err := fmt.Errorf("Cannot find hex gpu with pci address %s", pciAddress)
			log.Errorf(err.Error())
			return nil, err
		}

		hexProfilesMap := map[uint32]gpu.VgpuProfileFromHex{}
		hexProfileCollection := gpu.VgpuProfileCollectionFromHex{}

		if isVgpu(hexGpu) {
			hexProfilesMap, hexProfileCollection = cubecos.GetNodeVgpuProfilesMap(hexGpu.PciAddress)
		}

		attachedInstances := listAttachedInstances(listAttachedInstancesOpts{
			Device:                   device,
			DeviceUUID:               uuid,
			DeviceMemoryUsedMiB:      memoryUsedMiB,
			DeviceMemoryTotalMiB:     memoryTotalMiB,
			DeviceGpuUtilizationRate: gpuUtilizationPercent,
			NodeName:                 h.node,
			HexGpu:                   hexGpu,
			HexProfilesMap:           hexProfilesMap,
		})

		profileCollection := toProfileCollection(hexProfileCollection, attachedInstances)

		gpuCards = append(gpuCards, gpu.GpuCard{
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
			PciAddress: pciAddress,
			Status: gpu.GpuStatusInfo{
				Current:      hexGpu.Status,
				IsProcessing: false,
			},
			AllocationSummary: hexGpu.Allocation,
			ProfileCountLimit: hexGpu.ProfileCountLimit,
			Profiles:          profileCollection,
			AttachedInstances: attachedInstances,
		})
	}

	return gpuCards, nil
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

// Get PCI address in lowercase form.
func extractPciAddress(pciInfo nvml.PciInfo) string {
	// `busId` is a [32]int8 null-terminated string like "00000000:01:00.0"
	raw := pciInfo.BusId[:]
	buffer := make([]byte, len(raw))
	for i, b := range raw {
		buffer[i] = byte(b)
	}
	busId := strings.TrimRight(string(buffer), "\x00")
	address := strings.ToLower(strings.TrimSpace(busId))
	return address
}

func isVgpu(hexGpu gpu.GpuFromHex) bool {
	return hexGpu.Type == gpu.ResourceTypeSriovVgpu || hexGpu.Type == gpu.ResourceTypeMigBackedVgpu
}

func listAttachedInstances(opts listAttachedInstancesOpts) *[]gpu.AttachedInstance {
	hexGpu := opts.HexGpu

	switch opts.HexGpu.Type {
	case gpu.ResourceTypeUnset:
		return nil
	case gpu.ResourceTypePgpu:
		return listPgpuAttachedInstances(opts)
	case gpu.ResourceTypeSriovVgpu, gpu.ResourceTypeMigBackedVgpu:
		return listVgpuAttachedInstances(opts)
	default:
		log.Errorf("gpu: unhandled gpu type %s when listing attached instances for gpu %s", hexGpu.Type, hexGpu.Id)
	}

	return nil
}

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

	attachedInstance := cubecos.GetNodePgpuAttachedInstance(hexGpu.PciAddress)
	if attachedInstance == nil {
		log.Errorf("gpu: hex gpu %s allocation.current is not 0, but its attached instance is null on node %s", hexGpu.PciAddress, nodeName)
	} else {
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
	}

	return &attachedInstances
}

// Returns the attached instances for SR-IOV and MIG-backed vGPUs.
func listVgpuAttachedInstances(opts listAttachedInstancesOpts) *[]gpu.AttachedInstance {
	device, deviceUUID, hexProfilesMap := opts.Device, opts.DeviceUUID, opts.HexProfilesMap

	attachedInstances := []gpu.AttachedInstance{}
	vgpuInstances, ret := device.GetActiveVgpus()
	vgpuInstanceUtilizationMap := buildVgpuInstanceUtilizationMap(device, deviceUUID)

	if ret != nvml.SUCCESS {
		log.Errorf("nvml: failed to get active vgpus for device %s: %s", deviceUUID, nvml.ErrorString(ret))
		return &attachedInstances
	}

	openstackHelper := openstack.GetGlobalHelper()

	for i, vgpuInstance := range vgpuInstances {
		vmId, _, ret := vgpuInstance.GetVmID()
		if ret != nvml.SUCCESS {
			log.Errorf("nvml: failed to get VM ID for vgpu instance at index %d: %s", i, nvml.ErrorString(ret))
			continue
		}

		vgpuType, ret := vgpuInstance.GetType()
		if ret != nvml.SUCCESS {
			log.Errorf("nvml: failed to get type for vgpu instance %s: %s", vmId, nvml.ErrorString(ret))
			continue
		}

		profileId, ret := vgpuType.GetGpuInstanceProfileId()
		if ret != nvml.SUCCESS {
			log.Errorf("nvml: failed to get profile ID for vgpu instance %s: %s", vmId, nvml.ErrorString(ret))
			continue
		}

		frameBufferBytes, ret := vgpuType.GetFramebufferSize()
		if ret != nvml.SUCCESS {
			log.Errorf("nvml: failed to get frame buffer size for profile %d: %s", profileId, nvml.ErrorString(ret))
		}

		fbUsage, ret := vgpuInstance.GetFbUsage()
		if ret != nvml.SUCCESS {
			log.Errorf("nvml: failed to get fb usage for vgpu instance %s: %s", vmId, nvml.ErrorString(ret))
		}

		hexProfile := hexProfilesMap[profileId]
		profileAlias := hexProfile.Alias

		instanceName := ""
		server, err := openstackHelper.GetServer(vmId)
		if server == nil || err != nil {
			// Unable to get Openstack server is non-fatal.
			log.Errorf("gpu: failed to get Openstack server %s: %v", vmId, err)
		} else {
			instanceName = server.Name
		}

		vgpuInstanceId := uint32(reflect.ValueOf(vgpuInstance).Uint())
		utilizationPercent := vgpuInstanceUtilizationMap[vgpuInstanceId]

		attachedInstances = append(attachedInstances, gpu.AttachedInstance{
			Id:                 vmId,
			Name:               instanceName,
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

func buildVgpuInstanceUtilizationMap(device nvml.Device, deviceUUID string) map[uint32]uint32 {
	utilizationMap := map[uint32]uint32{}
	valueType, samples, ret := device.GetVgpuUtilization(0)

	if ret != nvml.SUCCESS {
		log.Errorf("nvml: failed to get vgpu utilization for device %s: %s", deviceUUID, nvml.ErrorString(ret))
		return utilizationMap
	}

	for _, sample := range samples {
		switch valueType {
		case nvml.VALUE_TYPE_UNSIGNED_INT:
			utilizationMap[sample.VgpuInstance] = binary.LittleEndian.Uint32(sample.SmUtil[:4])
		case nvml.VALUE_TYPE_DOUBLE:
			utilizationMap[sample.VgpuInstance] = binary.LittleEndian.Uint32(sample.SmUtil[:])
		}
	}

	return utilizationMap
}

func buildInstanceLinks(vmId string) gpu.InstanceLinks {
	grafanaLink := fmt.Sprintf(
		"https://%s/grafana/d/PVW6vU7Wz/instance?refresh=5m&kiosk=tv&orgId=1&var-UUID=%s",
		base.DataCenterVip,
		vmId,
	)

	openstackHelper := openstack.GetGlobalHelper()

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	result := remoteconsoles.Create(ctx, openstackHelper.Compute, vmId, remoteconsoles.CreateOpts{
		Protocol: remoteconsoles.ConsoleProtocolVNC,
		Type:     remoteconsoles.ConsoleTypeNoVNC,
	})
	console, err := result.Extract()

	consoleLink := ""

	if console == nil || err != nil {
		// Unable to create console link is non-fatal.
		log.Errorf("openstack: failed to create console link for instance %s: %v", vmId, err)
	} else {
		consoleLink = console.URL
	}

	return gpu.InstanceLinks{
		Grafana: grafanaLink,
		Console: consoleLink,
	}
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
