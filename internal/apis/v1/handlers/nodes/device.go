package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"k8s.io/client-go/util/workqueue"
)

var (
	lastDeviceList = sync.Map{}
	changes        = workqueue.NewTyped[nodes.Change]()
)

func (h *helper) listNodeDevices(opts nodes.DeviceListOpts) ([]nodes.BlockDevice, error) {
	if opts.Notify.Changes {
		defer cubecos.InsertNotification(opts.Notify.Payload)
	}

	if !opts.UseCache {
		return h.listNonCachedDevices()
	}

	devices, err := h.listCachedDevices()
	if err == nil {
		return devices, nil
	}

	return h.listNonCachedDevices()
}

func (h *helper) listCachedDevices() ([]nodes.BlockDevice, error) {
	cachedDevices, found := lastDeviceList.Load(h.node)
	if !found {
		err := fmt.Errorf("no cached devices found %s", h.node)
		log.Errorf("nodes(%s): %v", h.reqId, err)
		return nil, err
	}

	if len(cachedDevices.([]nodes.BlockDevice)) == 0 {
		err := fmt.Errorf("no cached devices found for node %s", h.node)
		log.Errorf("nodes(%s): %v", h.reqId, err)
		return nil, err
	}

	devices := cachedDevices.([]nodes.BlockDevice)
	h.syncUpdatingBlockDevices(&devices)
	return devices, nil
}

func (h *helper) listNonCachedDevices() ([]nodes.BlockDevice, error) {
	var err error
	blockDevs := []nodes.BlockDevice{}

	if nodes.IsLocal(h.node) {
		blockDevs, err = h.listLocalDevices()
	} else {
		blockDevs, err = h.listRemoteDevices()
	}
	if err != nil {
		log.Errorf("nodes: failed to list local devices for node %s(%v)", h.node, err)
		return nil, err
	}

	h.syncCachedBlockDevices(blockDevs)
	return blockDevs, nil
}

func (h *helper) listRemoteDevices() ([]nodes.BlockDevice, error) {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node(%s)(%v)", h.reqId, h.node, err)
		return nil, err
	}

	resp, err := h.http.R().
		SetResult(&devicesResp{}).
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.ListDevicesUrl())
	if err != nil {
		log.Errorf("nodes(%s): failed to list devices for node %s(%v)", h.reqId, h.node, err)
		return nil, err
	}

	if resp.IsError() {
		err := fmt.Errorf("failed to list devices for node %s(%s)", h.node, string(resp.Body()))
		log.Errorf("nodes(%s): %v", h.reqId, err)
		return nil, err
	}

	devResp := resp.Result().(*devicesResp)
	return devResp.Data, nil
}

func (h *helper) listLocalDevices() ([]nodes.BlockDevice, error) {
	raws, err := cubecos.GetRawBlockDevices()
	if err != nil {
		return nil, err
	}

	blockDevs := h.rawsToBlockDevices(raws)
	err = h.syncCephOsds(&blockDevs)
	if err != nil {
		return nil, err
	}

	h.syncUpdatingBlockDevices(&blockDevs)
	return blockDevs, nil
}

func (h *helper) rawsToBlockDevices(raws []nodes.RawBlockDevice) []nodes.BlockDevice {
	blockDevs := []nodes.BlockDevice{}
	mountsMap := map[string][]string{}

	for _, raw := range raws {
		if raw.IsPartition() {
			h.setPartitionMounts(mountsMap, raw)
			continue
		}

		blockDevs = append(
			blockDevs,
			cubecos.RawToBlockDevice(raw),
		)
	}

	h.setBlockDeviceAvailability(&blockDevs, mountsMap)
	h.setBlockDeviceStatus(&blockDevs)
	return blockDevs
}

func (h *helper) syncCephOsds(blockDevs *[]nodes.BlockDevice) error {
	cephDevs, err := ceph.GetDeviceMapByHost(h.node)
	if err != nil {
		log.Errorf("nodes: failed to list ceph devices by host %s(%v)", h.node, err)
		return err
	}

	for i, blockDev := range *blockDevs {
		cephDev, found := cephDevs[blockDev.Name]
		if !found {
			h.setOsdNotFoundInfo(&(*blockDevs)[i])
			continue
		}

		(*blockDevs)[i].Class = h.convergeClass(blockDev, cephDev.Osds)
		(*blockDevs)[i].Osd = nodes.Osd{
			Pgs:      h.getTotalPgs(cephDev.Osds),
			Reweigth: cephDev.Reweight,
			Daemons:  h.convertToOsds(cephDev.Osds),
		}
	}

	h.setPromotionDetails(blockDevs)
	return nil
}

func (h *helper) convergeClass(blockDev nodes.BlockDevice, osds []ceph.Osd) string {
	classes := []string{}
	for _, osd := range osds {
		if osd.DeviceClass != "" {
			classes = append(classes, strings.ToUpper(osd.DeviceClass))
		}
	}

	if len(classes) == 0 {
		return blockDev.Type
	}

	if slices.Contains(classes, blockdevice.SSD) {
		return blockdevice.SSD
	}

	return blockdevice.HDD
}

func (h *helper) getTotalPgs(osds []ceph.Osd) int {
	total := 0
	for _, osd := range osds {
		total += osd.Pgs
	}

	return total
}

func (h *helper) convertToOsds(list []ceph.Osd) []nodes.Deamon {
	deamons := []nodes.Deamon{}

	for _, osd := range list {
		deamons = append(
			deamons,
			nodes.Deamon{
				Id:           osd.Id,
				UsagePercent: osd.UsagePercent,
				Status:       status.Osd{Current: osd.Status},
			},
		)
	}

	return deamons
}

func (h *helper) setPartitionMounts(mountsMap map[string][]string, rawDev nodes.RawBlockDevice) {
	if rawDev.HasMountPoints() {
		mountsMap[rawDev.Name] = rawDev.MountPoints
	}
}

func (h *helper) setBlockDeviceAvailability(blockDevs *[]nodes.BlockDevice, mountsMap map[string][]string) {
	rawAvailabilities := h.genRawAvailabilities(mountsMap)
	devAvailabilities := h.converageDeviceAvailabilities(rawAvailabilities)

	for i, blockDev := range *blockDevs {
		availability, found := devAvailabilities[blockDev.Name]
		if found {
			(*blockDevs)[i].Availability = availability
		} else {
			(*blockDevs)[i].Availability = status.Available
		}
	}
}

func (h *helper) setBlockDeviceStatus(blockDevs *[]nodes.BlockDevice) {
	statusMap := h.syncBlockDeviceStatus(blockDevs)

	for i, blockDev := range *blockDevs {
		s, found := statusMap.Load(blockDev.Name)
		if found {
			(*blockDevs)[i].Status = s.(status.BlockDevice)
		}
	}
}

func (h *helper) setPromotionDetails(blockDevs *[]nodes.BlockDevice) {
	isCephHealthy := ceph.IsHealthy()

	for i, blockDev := range *blockDevs {
		if !isCephHealthy {
			(*blockDevs)[i].Status.IsPromotable = false
			(*blockDevs)[i].Status.IsDemotable = false
			(*blockDevs)[i].Status.Description = "ceph is not healthy, cannot promote or demote"
			continue
		}

		if strings.EqualFold(blockDev.Class, blockdevice.SSD) {
			(*blockDevs)[i].Status.IsPromotable = false
			(*blockDevs)[i].Status.IsDemotable = true
		} else {
			(*blockDevs)[i].Status.IsPromotable = true
			(*blockDevs)[i].Status.IsDemotable = false
		}
	}
}

func (h *helper) syncBlockDeviceStatus(blockDevs *[]nodes.BlockDevice) *sync.Map {
	wg := sync.WaitGroup{}
	statusMap := sync.Map{}

	for _, blockDev := range *blockDevs {
		wg.Add(1)
		go func(blockDev nodes.BlockDevice) {
			defer wg.Done()
			status := h.getBlockDeviceStatus(blockDev)
			statusMap.Store(blockDev.Name, status)
		}(blockDev)
	}

	wg.Wait()
	return &statusMap
}

func (h *helper) getBlockDeviceStatus(blockDev nodes.BlockDevice) status.BlockDevice {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(120))
	defer cancel()

	out, err := exec.
		CommandContext(ctx, "hex_sdk", "-f", "json", "ceph_osd_list", blockdevice.WithDevPath(blockDev.Name)).
		CombinedOutput()
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

func (h *helper) genRawAvailabilities(mountMap map[string][]string) map[string][]string {
	partAvailabilities := h.genPartitionAvailabilities(mountMap)
	return h.genDeviceAvailabilities(partAvailabilities)
}

func (h *helper) genPartitionAvailabilities(mountMap map[string][]string) map[string]string {
	availabilities := map[string]string{}
	for partition, paths := range mountMap {
		if len(paths) == 0 {
			continue
		}

		if paths[0] == "" {
			availabilities[partition] = status.Available
			continue
		}

		availabilities[partition] = status.InUse
		if h.isRootDevice(paths) {
			availabilities[partition] = status.System
		}
	}

	return availabilities
}

func (h *helper) isRootDevice(paths []string) bool {
	return slices.Contains(paths, "/")
}

func (h *helper) genDeviceAvailabilities(statuses map[string]string) map[string][]string {
	devStatuses := map[string][]string{}
	for partition, availability := range statuses {
		parent, err := h.getParentDev(partition)
		if err != nil {
			log.Errorf("nodes: failed to get parent device for %s(%v)", partition, err)
			continue
		}

		if parent == "" {
			log.Errorf("nodes: empty parent device for %s", partition)
			continue
		}

		devStatuses[parent] = append(
			devStatuses[parent],
			availability,
		)
	}

	return devStatuses
}

func (h *helper) converageDeviceAvailabilities(deviceAvailabilties map[string][]string) map[string]string {
	statuses := map[string]string{}
	for parent, availabilities := range deviceAvailabilties {
		if slices.Contains(availabilities, status.System) {
			statuses[parent] = status.System
			continue
		}

		if slices.Contains(availabilities, status.InUse) {
			statuses[parent] = status.InUse
			continue
		}

		statuses[parent] = status.Available
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
