package nodes

import (
	"regexp"

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
		addUptimeToNode(c, node, hypervisor)
	}
}

func addMetricToNode(node *definition.Node, hypervisor *hypervisors.Hypervisor) {
	node.Vcpu = definition.ComputeStatistic{
		TotalCores:  hypervisor.VCPUs,
		UsedCores:   hypervisor.VCPUsUsed,
		FreeCores:   hypervisor.VCPUs - hypervisor.VCPUsUsed,
		UsedPercent: float64(hypervisor.VCPUsUsed) / float64(hypervisor.VCPUs) * 100,
		FreePercent: float64(hypervisor.VCPUs-hypervisor.VCPUsUsed) / float64(hypervisor.VCPUs) * 100,
	}

	node.Memory = definition.SpaceStatistic{
		TotalMiB:    float64(hypervisor.MemoryMB),
		UsedMiB:     float64(hypervisor.MemoryMBUsed),
		FreeMiB:     float64(hypervisor.MemoryMB - hypervisor.MemoryMBUsed),
		UsedPercent: float64(hypervisor.MemoryMBUsed) / float64(hypervisor.MemoryMB) * 100,
		FreePercent: float64(hypervisor.MemoryMB-hypervisor.MemoryMBUsed) / float64(hypervisor.MemoryMB) * 100,
	}

	node.Storage = definition.SpaceStatistic{
		TotalMiB:    float64(hypervisor.LocalGB) * 1024,
		UsedMiB:     float64(hypervisor.LocalGBUsed) * 1024,
		FreeMiB:     float64(hypervisor.LocalGB-hypervisor.LocalGBUsed) * 1024,
		UsedPercent: float64(hypervisor.LocalGBUsed) / float64(hypervisor.LocalGB) * 100,
		FreePercent: float64(hypervisor.LocalGB-hypervisor.LocalGBUsed) / float64(hypervisor.LocalGB) * 100,
	}
}

func addUptimeToNode(c *gin.Context, node *definition.Node, hypervisor *hypervisors.Hypervisor) {
	h := openstack.GetGlobalHelper()
	time, err := h.GetHypervisorUpTime(hypervisor.ID)
	if err != nil {
		log.Debugf("request(%s): failed to add hypervisor uptime to the node: %s", api.GetReqId(c), err.Error())
		return
	}

	regex := regexp.MustCompile(`up\s+(.*?),`)
	match := regex.FindStringSubmatch(time.Uptime)
	if len(match) > 1 {
		node.Uptime = match[1]
		return
	}

	node.Uptime = "no uptime from system"
}
