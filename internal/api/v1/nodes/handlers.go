package nodes

import (
	"net/http"
	"regexp"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/nodes",
			Func:    getNodes,
		},
	}
)

func getNodes(c *gin.Context) {
	nodes, err := cubecos.ListNodes()
	if err != nil {
		//
		return
	}

	h, err := openstack.NewHelper(
		openstack.AuthType(config.Data.Spec.Openstack.Auth.Type),
		openstack.AuthUrl(config.Data.Spec.Openstack.Auth.Url),
		openstack.ProjectName(config.Data.Spec.Openstack.Auth.Project.Name),
		openstack.ProjectDomainName(config.Data.Spec.Openstack.Auth.Project.Domain.Name),
		openstack.Username(config.Data.Spec.Openstack.Auth.Username),
		openstack.Password(config.Data.Spec.Openstack.Auth.Password),
	)
	if err != nil {
		//
		return
	}

	licenses, err := cubecos.ListLicenses()
	if err != nil {
		//
		return
	}

	for i, node := range nodes {
		hypervisor, err := h.GetHypervisorByHostname(node.Hostname)
		if err != nil {
			//
			continue
		}

		nodes[i].ManagementIP = hypervisor.HostIP
		nodes[i].Status = hypervisor.State
		nodes[i].Vcpu = definition.ComputeStatistic{
			TotalCores: hypervisor.VCPUs,
			UsedCores:  hypervisor.VCPUsUsed,
			FreeCores:  hypervisor.VCPUs - hypervisor.VCPUsUsed,
		}
		nodes[i].Memory = definition.SpaceStatistic{
			TotalMiB: float64(hypervisor.MemoryMB),
			UsedMiB:  float64(hypervisor.MemoryMBUsed),
			FreeMiB:  float64(hypervisor.MemoryMB - hypervisor.MemoryMBUsed),
		}
		nodes[i].Storage = definition.SpaceStatistic{
			TotalMiB: float64(hypervisor.LocalGB) * 1024,
			UsedMiB:  float64(hypervisor.LocalGBUsed) * 1024,
			FreeMiB:  float64(hypervisor.LocalGB-hypervisor.LocalGBUsed) * 1024,
		}

		time, err := h.GetHypervisorUpTime(hypervisor.ID)
		if err != nil {
			//
			continue
		}

		re := regexp.MustCompile(`up\s+(.*?),`)
		match := re.FindStringSubmatch(time.Uptime)
		if len(match) > 1 {
			nodes[i].Uptime = match[1]
		} else {
			nodes[i].Uptime = "no uptime from system"
		}

		nodes[i].License = getLicenseByHostname(licenses, node.Hostname)
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"code":   http.StatusOK,
			"status": "ok",
			"msg":    "fetch nodes list successfully",
			"data":   nodes,
		},
	)
}

func getLicenseByHostname(licenses []definition.License, hostname string) definition.License {
	for _, license := range licenses {
		if license.Hostname == hostname {
			return license
		}
	}

	return definition.License{}
}
