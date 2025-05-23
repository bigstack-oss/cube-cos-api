package cubecos

import (
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
)

func GetVmStatus() (*VmStatus, error) {
	h := openstack.GetGlobalHelper()
	servers, err := h.ListServers(servers.ListOpts{AllTenants: true})
	if err != nil {
		log.Errorf("openstack: failed to list servers(%v)", err)
		return nil, err
	}

	return genVmStatusOverview(servers), nil
}

func GetVmUsage() (*metric.VmUsage, error) {
	h := openstack.GetGlobalHelper()
	stats, err := h.GetHypervisorStatistics()
	if err != nil {
		log.Errorf("openstack: failed to get hypervisor statistics(%v)", err)
		return nil, err
	}

	return genHypervisorUsage(stats), nil
}

func genVmStatusOverview(servers []servers.Server) *VmStatus {
	vm := &VmStatus{Total: len(servers)}

	for _, server := range servers {
		state := strings.ToLower(server.VmState)
		switch state {
		case "active":
			vm.Running++
		case "stopped":
			vm.Stopped++
		case "suspended":
			vm.Suspend++
		case "paused":
			vm.Paused++
		case "error":
			vm.Error++
		}
	}

	return vm
}

func genHypervisorUsage(stats *hypervisors.Statistics) *metric.VmUsage {
	return &metric.VmUsage{
		Vcpu: metric.Compute{
			TotalCores:  float64(stats.VCPUs),
			UsedCores:   float64(stats.VCPUsUsed),
			FreeCores:   float64(stats.VCPUs - stats.VCPUsUsed),
			UsedPercent: math.RoundDown((float64(stats.VCPUsUsed)/float64(stats.VCPUs))*100.0, 4),
			FreePercent: math.RoundDown((float64(stats.VCPUs-stats.VCPUsUsed)/float64(stats.VCPUs))*100.0, 4),
		},
		Memory: metric.Space{
			TotalMiB:    float64(stats.MemoryMB),
			UsedMiB:     float64(stats.MemoryMBUsed),
			FreeMiB:     float64(stats.MemoryMB - stats.MemoryMBUsed),
			UsedPercent: math.RoundDown((float64(stats.MemoryMBUsed)/float64(stats.MemoryMB))*100.0, 4),
			FreePercent: math.RoundDown((float64(stats.MemoryMB-stats.MemoryMBUsed)/float64(stats.MemoryMB))*100.0, 4),
		},
		Storage: GetCephUsage(),
	}
}
