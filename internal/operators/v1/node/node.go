package node

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/dustin/go-humanize"
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

func (o *Operator) Sync() {
	watcher, err := registry.Watch(
		registry.WatchService(definition.DataCenterName),
	)
	if err != nil {
		log.Errorf("failed to create watcher (%s)", err.Error())
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
	logThrottling(event)
}

func (o *Operator) setLicenseToNode(node *definition.Node) {
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

func (o *Operator) getLicenseByHostname(licenses []definition.License, hostname string) definition.License {
	for _, license := range licenses {
		if slices.Contains(license.Hosts, hostname) {
			license.Hosts = nil
			return license
		}
	}

	return definition.License{}
}

func (o *Operator) setInfraSpecToNode(node *definition.Node) {
	h := openstack.GetGlobalHelper()
	hypervisor, err := h.GetHypervisorByHostname(node.Hostname)
	if err != nil {
		log.Debugf("nodes: failed to add hypervisor info to the node: %s", err.Error())
		return
	}

	node.ManagementIP = node.Ip
	node.StorageIP = node.Ip
	node.Status = hypervisor.State
	o.addHardwareInfoToNode(node)
	o.addMetricToNode(node)
	o.addUptimeToNode(node)
}

func (o *Operator) addHardwareInfoToNode(node *definition.Node) {
	o.addCpuSpecToNode(node)
	o.addNetworkSpecToNode(node)
	o.addBlockDeviceSpecToNode(node)
}

func (o *Operator) addCpuSpecToNode(node *definition.Node) {
	info, err := cpu.Info()
	if err != nil {
		log.Errorf("nodes: failed to get cpu info: %s", err.Error())
		return
	}

	node.CpuSpec = info[0].ModelName
}

// M1 TODO: COS dev is working on the refactoring to support JSON output from 'hex_sdk DumpInterface'
// the implementation below will be replaced with the new one once it's ready
func (o *Operator) addNetworkSpecToNode(node *definition.Node) {
	out, err := exec.Command("hex_sdk", "DumpInterface").CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to get network info: %s", err.Error())
		return
	}

	node.NetworkInterfaces = []definition.NetworkInterface{}
	lines := strings.SplitSeq(string(out), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if isSkippableLine(line) {
			continue
		}

		fields := strings.Fields(line)
		if noEnoughNetFields(fields) {
			continue
		}

		node.NetworkInterfaces = append(
			node.NetworkInterfaces,
			definition.NetworkInterface{
				Label:       fields[0],
				BusIdSlaves: fields[1],
				Driver:      fields[2],
				State:       fields[3],
				Speed:       fields[4],
			},
		)
	}
}

func noEnoughNetFields(fields []string) bool {
	return len(fields) < 5
}

func (o *Operator) addBlockDeviceSpecToNode(node *definition.Node) {
	rawBlockDevs, err := getOsBlockDevices()
	if err != nil {
		return
	}

	node.BlockDevices = []definition.BlockDevice{}
	parentBlockDevs := map[string]string{}
	for _, rawBlockDev := range rawBlockDevs {
		if rawBlockDev.IsMainBlockDevice() {
			parentBlockDevs[rawBlockDev.Name] = rawBlockDev.Serial
			continue
		}

		node.BlockDevices = append(
			node.BlockDevices,
			convertToBlockDevice(rawBlockDev),
		)
	}

	addSerialToBlockDevices(node, parentBlockDevs)
}

func getOsBlockDevices() ([]definition.RawBlockDevice, error) {
	b, err := exec.Command("/bin/lsblk", "--sort", "name", "--json", "-o", "NAME,ROTA,SERIAL,SIZE,MOUNTPOINTS").Output()
	if err != nil {
		log.Errorf("nodes: failed to get block device info: %s", err.Error())
		return nil, err
	}

	blockDevMap := map[string][]definition.RawBlockDevice{}
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

func convertToBlockDevice(rawBlockDev definition.RawBlockDevice) definition.BlockDevice {
	return definition.BlockDevice{
		Serial:  rawBlockDev.Serial,
		Name:    rawBlockDev.Name,
		Type:    convertBlockDeviceType(rawBlockDev.Rota),
		SizeMiB: convertBlockDeviceSize(rawBlockDev.Size),
		Status:  identifyBlockDeviceStatus(rawBlockDev.MountPoints),
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

func identifyBlockDeviceStatus(mountPoints []string) string {
	if isNotMounted(mountPoints) {
		return "can be added"
	}

	return "in-use"
}

func addSerialToBlockDevices(node *definition.Node, parentBlockDevs map[string]string) {
	for name, serial := range parentBlockDevs {
		for i := range node.BlockDevices {
			if strings.Contains(node.BlockDevices[i].Name, name) {
				node.BlockDevices[i].Serial = serial
			}
		}
	}
}

func isNotMounted(mountPoints []string) bool {
	return len(mountPoints) == 0 || mountPoints[0] == ""
}

func isSkippableLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "Label")
}

func (o *Operator) addMetricToNode(node *definition.Node) {
	cpu, err := cubecos.GetCpuSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to get cpu summary of host: %s", err.Error())
	}
	if cpu == nil {
		cpu = &definition.ComputeStatistic{}
	}

	memory, err := cubecos.GetMemoryUsageSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to get memory summary of host: %s", err.Error())
	}
	if memory == nil {
		memory = &definition.SpaceStatistic{}
	}

	storage, err := cubecos.GetDiskStorageSummaryOfHost()
	if err != nil {
		log.Errorf("nodes: failed to get disk summary of host: %s", err.Error())
	}
	if storage == nil {
		storage = &definition.SpaceStatistic{}
	}

	node.Vcpu = *cpu
	node.Memory = *memory
	node.Storage = *storage
}

func (o *Operator) addUptimeToNode(node *definition.Node) {
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
