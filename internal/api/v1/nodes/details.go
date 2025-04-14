package nodes

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/cpu"
	log "go-micro.dev/v5/logger"
)

func (h *helper) addMetricsToNode(node *definition.Node) {
	openstack := openstack.GetGlobalHelper()
	hypervisor, err := openstack.GetHypervisorByHostname(node.Hostname)
	if err != nil {
		log.Debugf("nodes(%s): failed to add hypervisor info to the node: %s", api.GetReqId(h.c), err.Error())
		return
	}

	node.ManagementIP = definition.MgmtIP
	node.StorageIP = definition.StorageIP
	node.Status = hypervisor.State
	h.addHardwareInfoToNode(node)
	h.addMetricToNode(node)
	h.addUptimeToNode(node)
}

func (h *helper) addHardwareInfoToNode(node *definition.Node) {
	addCpuSpecToNode(node)
	addNetworkSpecToNode(node)
	addBlockDeviceSpecToNode(node)
}

func addCpuSpecToNode(node *definition.Node) {
	info, err := cpu.Info()
	if err != nil {
		log.Errorf("nodes: failed to get cpu info: %s", err.Error())
		return
	}

	node.CpuSpec = info[0].ModelName
}

// M1 TODO: COS dev is working on the refactoring to support JSON output from 'hex_sdk DumpInterface'
// the implementation below will be replaced with the new one once it's ready
func addNetworkSpecToNode(node *definition.Node) {
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

func addBlockDeviceSpecToNode(node *definition.Node) {
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

func (h *helper) addDetailsToNodes(nodes *[]definition.Node) {
	openstack := openstack.GetGlobalHelper()
	for i, node := range *nodes {
		hypervisor, err := openstack.GetHypervisorByHostname(node.Hostname)
		if err != nil {
			log.Debugf("nodes(%s): failed to add hypervisor info to the node: %s", api.GetReqId(h.c), err.Error())
			continue
		}

		(*nodes)[i].ManagementIP = definition.MgmtIP
		(*nodes)[i].StorageIP = definition.StorageIP
		(*nodes)[i].Status = hypervisor.State
		h.addHardwareInfoToNode((&(*nodes)[i]))
		h.addMetricToNode((&(*nodes)[i]))
		h.addUptimeToNode((&(*nodes)[i]))
	}
}

func (h *helper) addMetricToNode(node *definition.Node) {
	cpu, err := cubecos.GetCpuSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("nodes(%s): failed to get cpu summary of host: %s", api.GetReqId(h.c), err.Error())
	}
	if cpu == nil {
		cpu = &definition.ComputeStatistic{}
	}

	memory, err := cubecos.GetMemoryUsageSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("nodes(%s): failed to get memory summary of host: %s", api.GetReqId(h.c), err.Error())
	}
	if memory == nil {
		memory = &definition.SpaceStatistic{}
	}

	storage, err := cubecos.GetDiskStorageSummaryOfHost()
	if err != nil {
		log.Errorf("nodes(%s): failed to get disk summary of host: %s", api.GetReqId(h.c), err.Error())
	}
	if storage == nil {
		storage = &definition.SpaceStatistic{}
	}

	node.Vcpu = *cpu
	node.Memory = *memory
	node.Storage = *storage
}

func (h *helper) addUptimeToNode(node *definition.Node) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Errorf("nodes(%s): failed to read uptime file: %s", api.GetReqId(h.c), err.Error())
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		log.Errorf("nodes(%s): invalid uptime format", api.GetReqId(h.c))
		return
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Errorf("nodes(%s): failed to parse uptime: %s", api.GetReqId(h.c), err.Error())
		return
	}

	node.UptimeSeconds = uptimeSeconds
}
