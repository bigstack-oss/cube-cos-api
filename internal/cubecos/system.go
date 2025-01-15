package cubecos

import (
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
)

func GetResourceMetrics() (Metrics, error) {
	h, err := openstack.NewHelper()
	if err != nil {
		return Metrics{}, err
	}

	stats, err := h.GetHypervisorStatistics()
	if err != nil {
		return Metrics{}, err
	}

	return genResourceMetrics(stats), nil
}

func genResourceMetrics(stats *hypervisors.Statistics) Metrics {
	return Metrics{
		Vcpu: ComputeStatistic{
			TotalCores: stats.VCPUs,
			UsedCores:  stats.VCPUsUsed,
			FreeCores:  stats.VCPUs - stats.VCPUsUsed,
		},
		Memory: SpaceStatistic{
			TotalMiB: float64(stats.MemoryMB),
			UsedMiB:  float64(stats.MemoryMBUsed),
			FreeMiB:  float64(stats.MemoryMB - stats.MemoryMBUsed),
		},
		Storage: SpaceStatistic{
			TotalMiB: float64(stats.LocalGB) * 1024,
			UsedMiB:  float64(stats.LocalGBUsed) * 1024,
			FreeMiB:  float64(stats.LocalGB-stats.LocalGBUsed) * 1024,
		},
	}
}
