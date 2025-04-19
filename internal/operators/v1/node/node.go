package node

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/dustin/go-humanize"
	json "github.com/json-iterator/go"
	"github.com/shirou/gopsutil/v4/cpu"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

var (
	module = "node"
)

func Name() string {
	return module
}

type Operator struct {
	ctx             context.Context
	cancel          context.CancelFunc
	isFirstTimeSync bool
	sync            sync.Mutex
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	o.ctx = ctx
	o.cancel = cancel
	o.isFirstTimeSync = true
	o.sync = sync.Mutex{}
	go o.traceNodeDetails()
	return nil
}

func (o *Operator) Run() {
	watcher, err := registry.Watch(
		registry.WatchService(v1.DataCenterName),
	)
	if err != nil {
		log.Errorf("nodes: failed to create watcher (%s)", err.Error())
		return
	}

	defer watcher.Stop()
	select {
	case <-o.ctx.Done():
		return
	default:
		o.watchAndSyncNodeRoles(&watcher)
	}
}

func (o *Operator) Stop() {
	o.cancel()
}

func (o *Operator) watchAndSyncNodeRoles(watcher *registry.Watcher) {
	event, err := (*watcher).Next()
	if err != nil {
		log.Errorf("nodes: failed to get service discovery event", err.Error())
		return
	}

	o.syncNodeDetails()
	cubelog.Throttle("node", genDiscoveryMsg(event))
}

func (o *Operator) setLicenseToNode(node *v1.Node) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to add license info to the nodes: %s", err.Error())
		return
	}

	node.License = o.getLicenseByHostname(
		licenses,
		node.Hostname,
	)
}

func (o *Operator) getLicenseByHostname(licenses []v1.License, hostname string) v1.License {
	for _, license := range licenses {
		if slices.Contains(license.Hosts, hostname) {
			license.Hosts = nil
			return license
		}
	}

	return v1.License{}
}

func (o *Operator) setInfraSpecToNode(node *v1.Node) {
	h := openstack.GetGlobalHelper()
	hypervisor, err := h.GetHypervisorByHostname(node.Hostname)
	if err != nil {
		log.Debugf("nodes: failed to add hypervisor info to the node: %s", err.Error())
		return
	}

	node.ManagementIP = node.Ip
	node.StorageIP = node.Ip
	node.Status = hypervisor.State
	o.setHardwareInfoToNode(node)
	o.setMetricToNode(node)
	o.setUptimeToNode(node)
}

func (o *Operator) setHardwareInfoToNode(node *v1.Node) {
	o.setCpuSpecToNode(node)
	o.setNetworkSpecToNode(node)
	o.setBlockDeviceSpecToNode(node)
}

func (o *Operator) setCpuSpecToNode(node *v1.Node) {
	info, err := cpu.Info()
	if err != nil {
		log.Errorf("nodes: failed to get cpu info: %s", err.Error())
		return
	}

	node.CpuSpec = info[0].ModelName
}

func (o *Operator) setNetworkSpecToNode(node *v1.Node) {
	out, err := exec.Command("hex_sdk", "-f", "json", "DumpInterface").CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to get network info: %s", err.Error())
		return
	}

	raws := []v1.RawNetworkInterface{}
	err = json.Unmarshal(out, &raws)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal network info: %s", err.Error())
		return
	}

	for _, raw := range raws {
		node.NetworkInterfaces = append(
			node.NetworkInterfaces,
			v1.NetworkInterface(raw),
		)
	}
}

func (o *Operator) setBlockDeviceSpecToNode(node *v1.Node) {
	rawBlockDevs, err := getOsBlockDevices()
	if err != nil {
		return
	}

	node.BlockDevices = []v1.BlockDevice{}
	partitionMounts := map[string][]string{}
	for _, raw := range rawBlockDevs {
		if raw.IsPartition() {
			setPartitionMounts(raw, partitionMounts)
			continue
		}

		node.BlockDevices = append(
			node.BlockDevices,
			convertToBlockDevice(raw),
		)
	}

	addStatusToBlockDevices(node, partitionMounts)
}

func setPartitionMounts(blockDev v1.RawBlockDevice, partitionMounts map[string][]string) {
	if blockDev.NoMountPoints() {
		return
	}

	partitionMounts[blockDev.Name] = blockDev.MountPoints
}

func (o *Operator) setMetricToNode(node *v1.Node) {
	cpu, err := cubecos.GetCpuSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to get cpu summary of host: %s", err.Error())
	}
	if cpu == nil {
		cpu = &v1.ComputeStatistic{}
	}

	memory, err := cubecos.GetMemoryUsageSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to get memory summary of host: %s", err.Error())
	}
	if memory == nil {
		memory = &v1.SpaceStatistic{}
	}

	storage, err := cubecos.GetDiskStorageSummaryOfHost()
	if err != nil {
		log.Errorf("nodes: failed to get disk summary of host: %s", err.Error())
	}
	if storage == nil {
		storage = &v1.SpaceStatistic{}
	}

	node.Vcpu = *cpu
	node.Memory = *memory
	node.Storage = *storage
}

func (o *Operator) setUptimeToNode(node *v1.Node) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Errorf("nodes: failed to read uptime file: %s", err.Error())
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		log.Errorf("nodes: invalid uptime format")
		return
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Errorf("nodes: failed to parse uptime: %s", err.Error())
		return
	}

	node.UptimeSeconds = uptimeSeconds
}

func getOsBlockDevices() ([]v1.RawBlockDevice, error) {
	b, err := exec.Command("/bin/lsblk", "--sort", "name", "--json", "-o", "NAME,ROTA,SERIAL,SIZE,MOUNTPOINTS", "-e", v1.NetBlockDeviceCode).Output()
	if err != nil {
		log.Errorf("nodes: failed to get block device info: %s", err.Error())
		return nil, err
	}

	blockDevMap := map[string][]v1.RawBlockDevice{}
	err = json.Unmarshal(b, &blockDevMap)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal block device info: %s", err.Error())
		return nil, err
	}

	rawBlockDevs, found := blockDevMap["blockdevices"]
	if !found {
		log.Errorf("nodes: failed to find block devices in the output")
		return nil, err
	}
	if len(rawBlockDevs) <= 0 {
		log.Errorf("nodes: no block device found")
		return nil, errors.New("no block device found")
	}

	return rawBlockDevs, nil
}

func convertToBlockDevice(rawBlockDev v1.RawBlockDevice) v1.BlockDevice {
	return v1.BlockDevice{
		Serial:  rawBlockDev.Serial,
		Name:    rawBlockDev.Name,
		Type:    convertBlockDeviceType(rawBlockDev.Rota),
		SizeMiB: convertBlockDeviceSize(rawBlockDev.Size),
		Status:  "can be added",
	}
}

func convertBlockDeviceType(rota bool) string {
	if rota {
		return "HDD"
	}

	return "SSD"
}

func convertBlockDeviceSize(sizeStr string) float64 {
	bytes, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		log.Errorf("nodes: failed to convert block device size: %s", err.Error())
		return 0
	}

	sizeMiB := float64(bytes) / (1024.0 * 1024.0)
	return math.RoundDown(sizeMiB, 4)
}

func parentDeviceSysfs(device string) (string, error) {
	link := "/sys/class/block/" + device
	target, err := filepath.EvalSymlinks(link)
	if err != nil {
		log.Errorf("nodes: failed to get parent device for %s: %s", device, err.Error())
		return "", err
	}

	return filepath.Base(filepath.Dir(target)), nil
}

func addStatusToBlockDevices(node *v1.Node, partitions map[string][]string) {
	partitionStatuses := genPartitionStatuses(partitions)
	mainBlockDevStatus := genMainBlockDevStatus(partitionStatuses)
	for i, blockDev := range node.BlockDevices {
		node.BlockDevices[i].Status = mainBlockDevStatus[blockDev.Name]
	}
}

func genPartitionStatuses(partitions map[string][]string) map[string]string {
	partitionStatuses := map[string]string{}

	for partition, mountPoints := range partitions {
		if len(mountPoints) == 0 {
			continue
		}

		partitionStatuses[partition] = "in-use"
		if slices.Contains(mountPoints, "/") {
			partitionStatuses[partition] = "system"
		}
	}

	return partitionStatuses
}

func genMainBlockDevStatus(partitionStatuses map[string]string) map[string]string {
	mainBlockDevStatus := map[string]string{}
	for partition, status := range partitionStatuses {
		parent, err := parentDeviceSysfs(partition)
		if err != nil {
			log.Errorf("nodes: failed to get parent device: %s", err.Error())
			continue
		}

		val, found := mainBlockDevStatus[parent]
		if !found {
			mainBlockDevStatus[parent] = status
			continue
		}

		if val == "system" {
			continue
		}

		mainBlockDevStatus[parent] = status
	}

	return mainBlockDevStatus
}

func genDiscoveryMsg(event *registry.Result) string {
	return fmt.Sprintf(
		"node(%s) role(%s) ip(%s) %s",
		event.Service.Nodes[0].Metadata["hostname"],
		event.Service.Nodes[0].Metadata["role"],
		event.Service.Nodes[0].Address,
		convertAction(event.Action),
	)
}

func convertAction(action string) string {
	switch action {
	case status.Create:
		return "joined"
	case status.Delete:
		return "left"
	}

	return action
}
