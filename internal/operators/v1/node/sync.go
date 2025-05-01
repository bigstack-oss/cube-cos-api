package node

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/dustin/go-humanize"
	json "github.com/json-iterator/go"
	"github.com/shirou/gopsutil/v4/cpu"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

func (o *Operator) traceNodeDetails() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			o.syncNodeDetails()
			time.Sleep(time.Second * 30)
		}
	}
}

func (o *Operator) syncNodeDetails() {
	o.sync.Lock()
	defer o.sync.Unlock()

	v1.SyncRoleNodes()
	activeNodes := v1.GetActiveNodeMap()
	sourceNodes, err := cubecos.GetSourceNodeMap()
	if err != nil {
		log.Errorf("nodes: failed to get source node map: %s", err.Error())
		return
	}

	nodes := o.syncNodeStatus(sourceNodes, activeNodes)
	o.setNodeDetails(&nodes)
	v1.SetNodeDetails(nodes)
}

func (o *Operator) syncNodeStatus(sourceNodes, activeNodes map[string]v1.Node) []v1.Node {
	nodes := []v1.Node{}

	for _, srcNode := range sourceNodes {
		node, found := activeNodes[srcNode.Hostname]
		if found {
			nodes = append(nodes, node)
			continue
		}

		o.setNodeOfflineInfo(&srcNode)
		nodes = append(nodes, srcNode)
	}

	return nodes
}

func (o *Operator) setNodeOfflineInfo(node *v1.Node) {
	node.DataCenter = v1.DataCenterName
	node.BlockDevices = []v1.BlockDevice{}
	node.NetworkInterfaces = []v1.NetworkInterface{}
	node.License = o.getLicenseByHostname(node.Hostname)
	node.Status = "down"
}

func (o *Operator) setNodeDetails(nodes *[]v1.Node) {
	if len(*nodes) == 0 {
		return
	}

	for i, node := range *nodes {
		if node.IsLocal() {
			o.setNodeLicense(&(*nodes)[i])
			o.setNodeInfraSpec(&(*nodes)[i])
			continue
		}

		if node.IsDown() {
			continue
		}

		n, err := o.askPeerNode(node)
		if err == nil {
			(*nodes)[i].ManagementIP = n.ManagementIP
			(*nodes)[i].StorageIP = n.StorageIP
			(*nodes)[i].Vcpu = n.Vcpu
			(*nodes)[i].Memory = n.Memory
			(*nodes)[i].Storage = n.Storage
			(*nodes)[i].CpuSpec = n.CpuSpec
			(*nodes)[i].NetworkInterfaces = n.NetworkInterfaces
			(*nodes)[i].BlockDevices = n.BlockDevices
			(*nodes)[i].License = n.License
			(*nodes)[i].Status = n.Status
			(*nodes)[i].UptimeSeconds = n.UptimeSeconds
		}
	}
}

func (o *Operator) askPeerNode(node v1.Node) (*v1.Node, error) {
	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&api.NodeData{}).
		SetHeaders(v1.GenNodeAuthHeaders()).
		Get(node.GetNodeDetailsUrl())
	if err != nil {
		log.Errorf("nodes: failed to get node details %s: %s", node.Hostname, err.Error())
		return nil, err
	}
	if resp.IsError() {
		err := fmt.Errorf("get error for node details %s: %d(%s)", node.Hostname, resp.StatusCode(), string(resp.Body()))
		log.Errorf("nodes: %v", err)
		return nil, err
	}

	return &resp.Result().(*api.NodeData).Data, nil
}

func (o *Operator) setNodeLicense(node *v1.Node) {
	node.License = o.getLicenseByHostname(node.Hostname)
}

func (o *Operator) getLicenseByHostname(hostname string) v1.License {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("nodes: failed to add license info to the nodes: %s", err.Error())
		return v1.License{}
	}

	for _, license := range licenses {
		if slices.Contains(license.Hosts, hostname) {
			license.Hosts = nil
			return license
		}
	}

	return v1.License{}
}

func (o *Operator) setNodeInfraSpec(node *v1.Node) {
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

	addBlockDevicesAvailability(node, partitionMounts)
	addBlockDevicesStatus(node)
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
		Status:  status.BlockDevice{Current: "can be added"},
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

func addBlockDevicesAvailability(node *v1.Node, partitions map[string][]string) {
	partitionStatuses := genPartitionAvailability(partitions)
	mainBlockDevStatus := genMainBlockDevStatus(partitionStatuses)
	for i, blockDev := range node.BlockDevices {
		node.BlockDevices[i].Availability = mainBlockDevStatus[blockDev.Name]
	}
}

func addBlockDevicesStatus(node *v1.Node) {
	for i, blockDev := range node.BlockDevices {
		node.BlockDevices[i].Status = getBlockDeviceStatus(blockDev)
	}
}

func getBlockDeviceStatus(blockDev v1.BlockDevice) status.BlockDevice {
	out, err := exec.Command("smartctl", "-a", fmt.Sprintf("/dev/%s", blockDev.Name), "--json").CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to get block device(%s) status: %s", blockDev.Name, err.Error())
		return status.BlockDevice{Current: "failed"}
	}

	smartCtl := v1.SmartCtl{}
	err = json.Unmarshal(out, &smartCtl)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal block device(%s) status: %s", blockDev.Name, err.Error())
		return status.BlockDevice{Current: "failed"}
	}

	return status.BlockDevice{
		Current: parseSmartCtlStatus(smartCtl),
	}
}

func parseSmartCtlStatus(smartCtl v1.SmartCtl) string {
	if smartCtl.SmartStatus.Passed {
		return "ok"
	}

	return "failed"
}

func genPartitionAvailability(partitions map[string][]string) map[string]string {
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
