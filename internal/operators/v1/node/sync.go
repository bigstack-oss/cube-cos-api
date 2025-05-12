package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/shirou/gopsutil/v4/cpu"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

func (o *Operator) periodicSyncNodes() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			o.syncNodes()
			wait.Seconds(30)
		}
	}
}

func (o *Operator) syncNodes() {
	o.sync.Lock()
	defer o.sync.Unlock()

	nodes.Sync()
	allNodes, err := o.syncNodesUpAndDown()
	if err != nil {
		log.Errorf("nodes: failed to sync nodes: %v", err)
		return
	}

	o.syncDetails(&allNodes)
	nodes.SetList(allNodes)
}

func (o *Operator) syncNodesUpAndDown() ([]nodes.Node, error) {
	ups := nodes.GetMap()
	upsAndDowns, err := cubecos.GetSourceNodeMap()
	if err != nil {
		log.Errorf("nodes: failed to get source node map: %v", err)
		return nil, err
	}

	nodes := []nodes.Node{}
	for _, node := range upsAndDowns {
		up, found := ups[node.Hostname]
		if found {
			nodes = append(nodes, up)
			continue
		}

		o.setNodeDown(&node)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (o *Operator) setNodeDown(node *nodes.Node) {
	node.DataCenter = base.DataCenterName
	node.BlockDevices = []nodes.BlockDevice{}
	node.NetworkInterfaces = []nodes.NetworkInterface{}
	node.License = o.getLicenseByHostname(node.Hostname)
	node.Status = status.Down
}

func (o *Operator) syncDetails(nodes *[]nodes.Node) {
	if len(*nodes) == 0 {
		return
	}

	for i, node := range *nodes {
		if node.IsLocal() {
			o.setLicense(&(*nodes)[i])
			o.setInfraSpec(&(*nodes)[i])
			continue
		}

		if node.IsDown() {
			continue
		}

		n, err := o.askPeerNode(node)
		if err == nil {
			(*nodes)[i].ManagementIP = n.ManagementIP
			(*nodes)[i].SerialNumber = n.SerialNumber
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

func (o *Operator) askPeerNode(node nodes.Node) (*nodes.Node, error) {
	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&bodies.Node{}).
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.GetNodeUrl())
	if err != nil {
		log.Errorf("nodes: failed to get node details %s: %v", node.Hostname, err)
		return nil, err
	}

	if !resp.IsError() {
		return &resp.Result().(*bodies.Node).Data, nil
	}

	err = fmt.Errorf("resp error for node details %s: %s", node.Hostname, string(resp.Body()))
	log.Errorf("nodes: %v", err)
	return nil, err
}

func (o *Operator) setLicense(node *nodes.Node) {
	node.License = o.getLicenseByHostname(node.Hostname)
}

func (o *Operator) getLicenseByHostname(hostname string) licenses.License {
	list, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("nodes: failed to add license info to the nodes: %v", err)
		return licenses.License{}
	}

	for _, license := range list {
		if slices.Contains(license.Hosts, hostname) {
			license.Hosts = nil
			return license
		}
	}

	return licenses.License{}
}

func (o *Operator) setInfraSpec(node *nodes.Node) {
	o.setIps(node)
	o.setStatus(node)
	o.setHardwareSpec(node)
	o.setMetric(node)
	o.setUptime(node)
}

func (o *Operator) setIps(node *nodes.Node) {
	node.ManagementIP = base.ManagementIp
	node.StorageIP = base.StorageIP
}

func (o *Operator) setStatus(n *nodes.Node) {
	switch base.CurrentRole {
	case nodes.RoleControl, nodes.RoleControlConverged, nodes.RoleModerator, nodes.RoleStorage:
		n.Status = status.Up
	case nodes.RoleCompute, nodes.RoleEdgeCore:
		o.setComputeStatus(n)
	}
}

func (o *Operator) setComputeStatus(node *nodes.Node) {
	h := openstack.GetGlobalHelper()
	hypervisor, err := h.GetHypervisorByHostname(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to add hypervisor info to the node: %v", err)
		return
	}

	node.Status = hypervisor.State
}

func (o *Operator) setHardwareSpec(node *nodes.Node) {
	o.setCpuSpec(node)
	o.setNetworkSpec(node)
	o.setStorageSpec(node)
}

func (o *Operator) setCpuSpec(node *nodes.Node) {
	info, err := cpu.Info()
	if err != nil {
		log.Errorf("nodes: failed to get cpu info: %v", err)
		return
	}

	node.CpuSpec = info[0].ModelName
}

func (o *Operator) setNetworkSpec(n *nodes.Node) {
	interfaces, err := cubecos.DumpInterfaces()
	if err != nil {
		return
	}

	for _, net := range interfaces {
		n.NetworkInterfaces = append(
			n.NetworkInterfaces,
			nodes.NetworkInterface(net),
		)
	}
}

func (o *Operator) setStorageSpec(n *nodes.Node) {
	raws, err := cubecos.GetRawBlockDevices()
	if err != nil {
		return
	}

	n.BlockDevices = []nodes.BlockDevice{}
	partitionMounts := map[string][]string{}
	for _, raw := range raws {
		if raw.IsPartition() {
			o.setPartitionMounts(raw, partitionMounts)
			continue
		}

		n.BlockDevices = append(
			n.BlockDevices,
			cubecos.ConvertToBlockDevice(raw),
		)
	}

	o.setBlockDevicesAvailability(n, partitionMounts)
	o.setBlockDevicesStatus(n)
}

func (o *Operator) setPartitionMounts(partition nodes.RawBlockDevice, mounts map[string][]string) {
	if partition.NoMountPoints() {
		return
	}

	mounts[partition.Name] = partition.MountPoints
}

func (o *Operator) setBlockDevicesAvailability(node *nodes.Node, partitions map[string][]string) {
	partitionStatuses := genPartitionAvailability(partitions)
	mainBlockDevStatus := genMainBlockDevStatus(partitionStatuses)
	for i, blockDev := range node.BlockDevices {
		node.BlockDevices[i].Availability = mainBlockDevStatus[blockDev.Name]
	}
}

func (o *Operator) setBlockDevicesStatus(node *nodes.Node) {
	for i, blockDev := range node.BlockDevices {
		node.BlockDevices[i].Status = getBlockDeviceStatus(blockDev)
	}
}

func (o *Operator) setMetric(node *nodes.Node) {
	if node.IsDown() {
		return
	}

	cpu, err := cubecos.GetHostCpuSummary(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to set cpu summary of host: %v", err)
	}
	if cpu == nil {
		cpu = &metric.Compute{}
	}

	memory, err := cubecos.GetHostMemoryUsageSummary(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to get memory summary of host: %v", err)
	}
	if memory == nil {
		memory = &metric.Space{}
	}

	storage, err := cubecos.GetHostDiskStorageSummary()
	if err != nil {
		log.Errorf("nodes: failed to get disk summary of host: %v", err)
	}
	if storage == nil {
		storage = &metric.Space{}
	}

	node.Vcpu = *cpu
	node.Memory = *memory
	node.Storage = *storage
}

func (o *Operator) setUptime(node *nodes.Node) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Errorf("nodes: failed to read uptime file: %v", err)
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		log.Errorf("nodes: invalid uptime format")
		return
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Errorf("nodes: failed to parse uptime: %v", err)
		return
	}

	node.UptimeSeconds = uptimeSeconds
}

func parentDeviceSysfs(device string) (string, error) {
	link := "/sys/class/block/" + device
	target, err := filepath.EvalSymlinks(link)
	if err != nil {
		log.Errorf("nodes: failed to get parent device for %s: %v", device, err)
		return "", err
	}

	return filepath.Base(filepath.Dir(target)), nil
}

func getBlockDeviceStatus(blockDev nodes.BlockDevice) status.BlockDevice {
	out, err := exec.Command("hex_sdk", "-f", "json", "ceph_osd_list", fmt.Sprintf("/dev/%s", blockDev.Name)).CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to get block device(%s) status: %v", blockDev.Name, err)
		return status.BlockDevice{Current: "failed"}
	}

	smartCtl := []blockdevice.SmartCtl{}
	err = json.Unmarshal(out, &smartCtl)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal block device(%s) status: %v", blockDev.Name, err)
		return status.BlockDevice{Current: "failed"}
	}

	return convergeBlockStatuses(smartCtl)
}

func convergeBlockStatuses(statuses []blockdevice.SmartCtl) status.BlockDevice {
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
			log.Errorf("nodes: failed to get parent device: %v", err)
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
