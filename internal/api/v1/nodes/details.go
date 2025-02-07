package nodes

import (
	"os"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	log "go-micro.dev/v5/logger"
)

func addNodeDetailsToNodes(c *gin.Context, nodes *[]*definition.Node) {
	h := openstack.GetGlobalHelper()
	for _, node := range *nodes {
		hypervisor, err := h.GetHypervisorByHostname(node.Hostname)
		if err != nil {
			log.Debugf("request(%s): failed to add hypervisor info to the node: %s", api.GetReqId(c), err.Error())
			continue
		}

		node.ManagementIP = hypervisor.HostIP
		node.Status = hypervisor.State
		addMetricToNode(node, hypervisor)
		addUptimeToNode(c, node)
	}
}

func addMetricToNode(node *definition.Node, hypervisor *hypervisors.Hypervisor) {
	node.Vcpu = definition.ComputeStatistic{
		TotalCores:  hypervisor.VCPUs,
		UsedCores:   hypervisor.VCPUsUsed,
		FreeCores:   hypervisor.VCPUs - hypervisor.VCPUsUsed,
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

func addUptimeToNode(c *gin.Context, node *definition.Node) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Errorf("request(%s): failed to read uptime file: %s", api.GetReqId(c), err.Error())
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		log.Errorf("request(%s): invalid uptime format", api.GetReqId(c))
		return
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Errorf("request(%s): failed to parse uptime: %s", api.GetReqId(c), err.Error())
		return
	}

	node.UptimeSeconds = uptimeSeconds
}
