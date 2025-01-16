package cubecos

import (
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
)

func GetResourceMetrics() (definition.Metrics, error) {
	h, err := openstack.NewHelper()
	if err != nil {
		return definition.Metrics{}, err
	}

	stats, err := h.GetHypervisorStatistics()
	if err != nil {
		return definition.Metrics{}, err
	}

	return genResourceMetrics(stats), nil
}

func genResourceMetrics(stats *hypervisors.Statistics) definition.Metrics {
	return definition.Metrics{
		Vcpu: definition.ComputeStatistic{
			TotalCores: stats.VCPUs,
			UsedCores:  stats.VCPUsUsed,
			FreeCores:  stats.VCPUs - stats.VCPUsUsed,
		},
		Memory: definition.SpaceStatistic{
			TotalMiB: float64(stats.MemoryMB),
			UsedMiB:  float64(stats.MemoryMBUsed),
			FreeMiB:  float64(stats.MemoryMB - stats.MemoryMBUsed),
		},
		Storage: definition.SpaceStatistic{
			TotalMiB: float64(stats.LocalGB) * 1024,
			UsedMiB:  float64(stats.LocalGBUsed) * 1024,
			FreeMiB:  float64(stats.LocalGB-stats.LocalGBUsed) * 1024,
		},
	}
}
