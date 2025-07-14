package nodes

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) listNodeDevices() ([]nodes.BlockDevice, error) {
	rawDevs, err := cubecos.GetRawBlockDevices()
	if err != nil {
		log.Errorf("nodes: failed to get raw block devices(%v)", err)
		return nil, err
	}

	blockDevs := h.convertToBlockDevices(rawDevs)
	return blockDevs, nil
}

func (h *helper) convertToBlockDevices(rawDevs []nodes.RawBlockDevice) []nodes.BlockDevice {
	blockDevs := []nodes.BlockDevice{}
	mountsMap := map[string][]string{}

	for _, rawDev := range rawDevs {
		if rawDev.IsPartition() {
			h.setPartitionMounts(mountsMap, rawDev)
			continue
		}

		blockDevs = append(
			blockDevs,
			cubecos.ConvertToBlockDevice(rawDev),
		)
	}

	h.setBlockDevicesAvailability(&blockDevs, mountsMap)
	h.setBlockDevicesStatus(&blockDevs)
	return blockDevs
}

func (h *helper) setPartitionMounts(mountsMap map[string][]string, rawDev nodes.RawBlockDevice) {
	if rawDev.HasMountPoints() {
		mountsMap[rawDev.Name] = rawDev.MountPoints
	}
}

func (h *helper) setBlockDevicesAvailability(blockDevs *[]nodes.BlockDevice, mountsMap map[string][]string) {
	partitionAvailability := h.genPartitionAvailability(mountsMap)
	parentDevAvailability := h.genParentDevAvailability(partitionAvailability)
	for i, blockDev := range *blockDevs {
		(*blockDevs)[i].Availability = parentDevAvailability[blockDev.Name]
	}
}

func (h *helper) setBlockDevicesStatus(blockDevs *[]nodes.BlockDevice) {
	for i, blockDev := range *blockDevs {
		(*blockDevs)[i].Status = h.getBlockDeviceStatus(blockDev)
	}
}

func (h *helper) getBlockDeviceStatus(blockDev nodes.BlockDevice) status.BlockDevice {
	out, err := exec.Command("hex_sdk", "-f", "json", "ceph_osd_list", fmt.Sprintf("/dev/%s", blockDev.Name)).CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to get block device(%s) status(%v)", blockDev.Name, err)
		return status.BlockDevice{Current: "failed"}
	}

	smartCtl := []blockdevice.SmartCtl{}
	err = json.Unmarshal(out, &smartCtl)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal block device(%s) status(%v)", blockDev.Name, err)
		return status.BlockDevice{Current: "failed"}
	}

	return h.convergeBlockStatuses(smartCtl)
}

func (h *helper) convergeBlockStatuses(statuses []blockdevice.SmartCtl) status.BlockDevice {
	status := status.BlockDevice{Current: "ok"}

	for _, s := range statuses {
		if s.State != "ok" {
			status.Current = s.State
			status.Description = s.Remark
			return status
		}
	}

	return status
}

func (h *helper) genPartitionAvailability(mountMap map[string][]string) map[string]string {
	statuses := map[string]string{}

	for partition, paths := range mountMap {
		if len(paths) == 0 {
			continue
		}

		statuses[partition] = status.InUse
		if h.isRootDevice(paths) {
			statuses[partition] = status.System
		}
	}

	return statuses
}

func (h *helper) isRootDevice(paths []string) bool {
	return slices.Contains(paths, "/")
}

func (h *helper) genParentDevAvailability(partitionAvailability map[string]string) map[string]string {
	statuses := map[string]string{}
	for partition, availability := range partitionAvailability {
		parent, err := h.getParentDev(partition)
		if err != nil {
			log.Errorf("nodes: failed to get parent device(%v)", err)
			continue
		}

		val, found := statuses[parent]
		if !found {
			statuses[parent] = availability
			continue
		}

		if val == status.System {
			continue
		}

		statuses[parent] = availability
	}

	return statuses
}

func (h *helper) getParentDev(device string) (string, error) {
	link := "/sys/class/block/" + device
	target, err := filepath.EvalSymlinks(link)
	if err != nil {
		log.Errorf("nodes: failed to get parent device for %s(%v)", device, err)
		return "", err
	}

	return filepath.Base(
		filepath.Dir(target),
	), nil
}
