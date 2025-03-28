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
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/dustin/go-humanize"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
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

	node.ManagementIP = hypervisor.HostIP
	node.Status = hypervisor.State
	h.addHardwareInfoToNode(node)
	h.addMetricToNode(node, hypervisor)
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

	for _, rawBlockDev := range rawBlockDevs {
		node.BlockDevices = append(
			node.BlockDevices,
			definition.BlockDevice{
				Name:    rawBlockDev.Name,
				Type:    rawBlockDev.Type,
				SizeMiB: convertBlockDeviceSize(rawBlockDev.Size),
				Status:  identifyBlockDeviceStatus(rawBlockDev.MountPoints),
			},
		)
	}
}

func getOsBlockDevices() ([]definition.RawBlockDevice, error) {
	b, err := exec.Command("/bin/lsblk", "--sort", "name", "--json").Output()
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
		return "available"
	}

	return "storage"
}

func isNotMounted(mountPoints []string) bool {
	return len(mountPoints) == 0 || mountPoints[0] == ""
}

func isSkippableLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "Label")
}

func (h *helper) addDetailsToNodes(nodes *[]*definition.Node) {
	openstack := openstack.GetGlobalHelper()
	for _, node := range *nodes {
		hypervisor, err := openstack.GetHypervisorByHostname(node.Hostname)
		if err != nil {
			log.Debugf("request(%s): failed to add hypervisor info to the node: %s", api.GetReqId(h.c), err.Error())
			continue
		}

		node.ManagementIP = hypervisor.HostIP
		node.Status = hypervisor.State
		h.addHardwareInfoToNode(node)
		h.addMetricToNode(node, hypervisor)
		h.addUptimeToNode(node)
	}
}

func (h *helper) addMetricToNode(node *definition.Node, hypervisor *hypervisors.Hypervisor) {
	node.Vcpu = definition.ComputeStatistic{
		TotalCores:  float64(hypervisor.VCPUs),
		UsedCores:   float64(hypervisor.VCPUsUsed),
		FreeCores:   float64(hypervisor.VCPUs - hypervisor.VCPUsUsed),
		UsedPercent: math.RoundDown(float64(hypervisor.VCPUsUsed)/float64(hypervisor.VCPUs)*100, 4),
		FreePercent: math.RoundDown(float64(hypervisor.VCPUs-hypervisor.VCPUsUsed)/float64(hypervisor.VCPUs)*100, 4),
	}

	node.Memory = definition.SpaceStatistic{
		TotalMiB:    float64(hypervisor.MemoryMB),
		UsedMiB:     float64(hypervisor.MemoryMBUsed),
		FreeMiB:     float64(hypervisor.MemoryMB - hypervisor.MemoryMBUsed),
		UsedPercent: math.RoundDown(float64(hypervisor.MemoryMBUsed)/float64(hypervisor.MemoryMB)*100, 4),
		FreePercent: math.RoundDown(float64(hypervisor.MemoryMB-hypervisor.MemoryMBUsed)/float64(hypervisor.MemoryMB)*100, 4),
	}

	node.Storage = definition.SpaceStatistic{
		TotalMiB:    float64(hypervisor.LocalGB) * 1024,
		UsedMiB:     float64(hypervisor.LocalGBUsed) * 1024,
		FreeMiB:     float64(hypervisor.LocalGB-hypervisor.LocalGBUsed) * 1024,
		UsedPercent: math.RoundDown(float64(hypervisor.LocalGBUsed)/float64(hypervisor.LocalGB)*100, 4),
		FreePercent: math.RoundDown(float64(hypervisor.LocalGB-hypervisor.LocalGBUsed)/float64(hypervisor.LocalGB)*100, 4),
	}
}

func (h *helper) addUptimeToNode(node *definition.Node) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Errorf("request(%s): failed to read uptime file: %s", api.GetReqId(h.c), err.Error())
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		log.Errorf("request(%s): invalid uptime format", api.GetReqId(h.c))
		return
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Errorf("request(%s): failed to parse uptime: %s", api.GetReqId(h.c), err.Error())
		return
	}

	node.UptimeSeconds = uptimeSeconds
}
