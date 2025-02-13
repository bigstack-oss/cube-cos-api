package cubecos

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
)

func GetVmStatus() (*VmStatus, error) {
	h := openstack.GetGlobalHelper()
	servers, err := h.ListServers(servers.ListOpts{AllTenants: true})
	if err != nil {
		log.Errorf("failed to list servers: %v", err)
		return nil, err
	}

	return genVmStatusOverview(servers), nil
}

func GetVmUsage() (*definition.VmUsage, error) {
	h := openstack.GetGlobalHelper()
	stats, err := h.GetHypervisorStatistics()
	if err != nil {
		return nil, err
	}

	return genHypervisorUsage(stats), nil
}

func genVmStatusOverview(servers []servers.Server) *VmStatus {
	vm := &VmStatus{Total: len(servers)}

	for _, server := range servers {
		switch server.PowerState.String() {
		case "RUNNING":
			vm.Running++
		case "SHUTDOWN":
			vm.Stopped++
		case "SUSPENDED":
			vm.Suspend++
		case "PAUSED":
			vm.Paused++
		case "CRASHED":
			vm.Error++
		default:
			vm.Unknown++
		}
	}

	return vm
}

func genHypervisorUsage(stats *hypervisors.Statistics) *definition.VmUsage {
	return &definition.VmUsage{
		Vcpu: definition.ComputeStatistic{
			TotalCores:  float64(stats.VCPUs),
			UsedCores:   float64(stats.VCPUsUsed),
			FreeCores:   float64(stats.VCPUs - stats.VCPUsUsed),
			UsedPercent: math.RoundDown(float64(stats.VCPUsUsed)/float64(stats.VCPUs), 4),
			FreePercent: math.RoundDown(float64(stats.VCPUs-stats.VCPUsUsed)/float64(stats.VCPUs), 4),
		},
		Memory: definition.SpaceStatistic{
			TotalMiB:    float64(stats.MemoryMB),
			UsedMiB:     float64(stats.MemoryMBUsed),
			FreeMiB:     float64(stats.MemoryMB - stats.MemoryMBUsed),
			UsedPercent: math.RoundDown(float64(stats.MemoryMBUsed)/float64(stats.MemoryMB), 4),
			FreePercent: math.RoundDown(float64(stats.MemoryMB-stats.MemoryMBUsed)/float64(stats.MemoryMB), 4),
		},
		Storage: definition.SpaceStatistic{
			TotalMiB:    float64(stats.LocalGB) * 1024,
			UsedMiB:     float64(stats.LocalGBUsed) * 1024,
			FreeMiB:     float64(stats.LocalGB-stats.LocalGBUsed) * 1024,
			UsedPercent: math.RoundDown(float64(stats.LocalGBUsed)/float64(stats.LocalGB), 4),
			FreePercent: math.RoundDown(float64(stats.LocalGB-stats.LocalGBUsed)/float64(stats.LocalGB), 4),
		},
	}
}
