package nodes

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/remoteconsoles"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
)

type listAttachedInstancesOpts struct {
	Device nvml.Device
	DeviceUUID string
	DeviceMemoryUsedMiB int
	DeviceMemoryTotalMiB int
	DeviceGpuUtilizationRate uint32
	NodeName string
	HexGpu gpu.GpuFromHex
	HexProfilesMap map[uint32]gpu.VgpuProfileFromHex
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

		vgpuProfiles, hexProfilesMap := listVgpuProfiles(device, hexGpu)
		attachedInstances := listAttachedInstances(listAttachedInstancesOpts{
			Device: device,
			DeviceUUID: uuid,
			DeviceMemoryUsedMiB: memoryUsedMiB,
			DeviceMemoryTotalMiB: memoryTotalMiB,
			DeviceGpuUtilizationRate: gpuUtilizationPercent,
			NodeName: h.node,
			HexGpu: hexGpu,
			HexProfilesMap: hexProfilesMap,
		})

		if vgpuProfiles != nil && attachedInstances != nil {
			updateVgpuProfilesRemaining(*vgpuProfiles, *attachedInstances)
		}

		gpuCards = append(gpuCards, gpu.GpuCard{
			Id: hexGpu.Id,
			Name: hexGpu.Name,
			ResourceType: hexGpu.Type,
			Vram: &gpu.VramInfo{
				AllocatedMiB: memoryUsedMiB,
				TotalMiB: memoryTotalMiB,
				UtilizationPercent: float64(memoryUtilizationPercent),
			},
			Gpu: &gpu.GpuInfo{
				UtilizationPercent: float64(gpuUtilizationPercent),
			},
			PciAddress: pciAddress,
			Status: gpu.GpuStatusInfo{
				Current: hexGpu.Status,
				IsProcessing: false,
			},
			AllocationSummary: &gpu.AllocationSummary{
				Current: hexGpu.Allocation.Current,
				Total: hexGpu.Allocation.Total,
			},
			Profiles: vgpuProfiles,
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

// Get PCI address in lowercase 4-digit domain form.
// Handles both "0000:01:00.0" (Cyborg) and "00000000:01:00.0" (NVML BusId).
func extractPciAddress(pciInfo nvml.PciInfo) string {
	// `busId` is a [32]uint8 null-terminated string like "00000000:01:00.0"
	busId := strings.TrimRight(string(pciInfo.BusId[:]), "\x00")
	address := strings.ToLower(strings.TrimSpace(busId))
	parts := strings.SplitN(address, ":", 2)

	if len(parts) == 2 && len(parts[0]) == 8 {
		return parts[0][4:] + ":" + parts[1]
	}

	return address
}

// Returns vGPU profiles with `Remaining = Count`, or returns `nil` for non-vGPU.
// The exact `Remaining` value should be calculated based on the profile's `Count`
// and the amount of attached instances created with this profile.
func listVgpuProfiles(device nvml.Device, hexGpu gpu.GpuFromHex) (*[]gpu.VgpuProfile, map[uint32]gpu.VgpuProfileFromHex) {
	hexProfilesMap := map[uint32]gpu.VgpuProfileFromHex{}

	if !isVgpu(hexGpu) {
		return nil, hexProfilesMap
	}

	vgpuProfiles := []gpu.VgpuProfile{}
	hexProfilesMap = cubecos.GetNodeVgpuProfilesMap(hexGpu.PciAddress)

	for i := 0; ; i++ {
		nvmlProfile, ret := device.GetGpuInstanceProfileInfo(i)

		if ret == nvml.ERROR_INVALID_ARGUMENT {
			// No more profiles.
			break
		}

		if ret != nvml.SUCCESS {
			log.Errorf("nvml: failed to get gpu instance profile info for gpu %s at index %d: %v", hexGpu.Id, i, nvml.ErrorString(ret))
			continue
		}

		hexProfile := hexProfilesMap[nvmlProfile.Id]

		vgpuProfiles = append(vgpuProfiles, gpu.VgpuProfile{
			Id: strconv.FormatUint(uint64(nvmlProfile.Id), 10),
			Name: hexProfile.Name,
			VramMiB: nvmlProfile.MemorySizeMB,
			AliasName: hexProfile.Alias,
			Count: hexProfile.Count,
			// `Remaining` will be calculated later after getting attached instances.
			Remaining: hexProfile.Count,
		})
	}

	return &vgpuProfiles, hexProfilesMap
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
	case gpu.ResourceTypeSriovVgpu:
	case gpu.ResourceTypeMigBackedVgpu:
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

	if hexGpu.Allocation.Current == 0 {
		return &attachedInstances
	}

	activeOpenstackPgpuServer := getActiveOpenstackPgpuServer(nodeName)
	if activeOpenstackPgpuServer == nil {
		log.Errorf("gpu: hex pgpu allocation.current is not 0, but cannot find any active Openstack pgpu server on node %s", nodeName)
	} else {
		attachedInstances = append(attachedInstances, gpu.AttachedInstance{
			Id: activeOpenstackPgpuServer.ID,
			Name: activeOpenstackPgpuServer.Name,
			ProfileAlias: nil,
			UtilizationPercent: deviceGpuUtilizationRate,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: deviceMemoryUsedMiB,
				TotalMiB: deviceMemoryTotalMiB,
			},
			Links: buildInstanceLinks(activeOpenstackPgpuServer.ID),
		})
	}

	return &attachedInstances
}

func getActiveOpenstackPgpuServer(nodeName string) *servers.Server {
	openstackHelper := openstack.GetGlobalHelper()

	serversOnNode, err := openstackHelper.ListServers(servers.ListOpts{
		Host: nodeName,
		AllTenants: true,
	})

	if err != nil {
		log.Errorf("openstack: failed to list servers for node %s: %v", nodeName, err)
		return nil
	}

	for _, server := range serversOnNode {
		if server.Status != "ACTIVE" {
			continue
		}
		// TODO: Check `server.Flavor.ExtraSpecs` for "pci_passthrough:alias" key (non-empty = PCI passthrough VM)
		return &server
	}

	return nil
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
			Id: vmId,
			Name: instanceName,
			ProfileAlias: &profileAlias,
			UtilizationPercent: utilizationPercent,
			MemoryUsage: gpu.InstanceMemoryUsage{
				AllocatedMiB: bytesToMiB(fbUsage),
				TotalMiB: bytesToMiB(frameBufferBytes),
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

func updateVgpuProfilesRemaining(profiles []gpu.VgpuProfile, attachedInstances []gpu.AttachedInstance) {
	profileMapByAlias := map[string]gpu.VgpuProfile{}
	for _, profile := range profiles {
		profileMapByAlias[profile.AliasName] = profile
	}

	profileInstanceCountMap := map[string]int{}
	for _, instance := range attachedInstances {
		profile := profileMapByAlias[*instance.ProfileAlias]
		profileInstanceCountMap[profile.Id]++
	}

	for _, profile := range profiles {
		instanceCount := profileInstanceCountMap[profile.Id]
		profile.Remaining = profile.Count - instanceCount
	}
}