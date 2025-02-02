package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getCpuMetrics() (interface{}, error) {
	switch h.resourceType {
	case "hosts":
		return h.getHostCpuMetrics()
	case "vms":
		return h.getVmCpuMetrics()
	}

	return nil, fmt.Errorf(
		"invalid resource type(%s) to get cpu metrics",
		h.resourceType,
	)
}

func (h *helper) getHostCpuMetrics() (interface{}, error) {
	switch h.reportType {
	case "summary":
		return cubecos.GetHostCpuSummary()
	case "rank":
		return cubecos.GetHostCpuRank(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get host cpu metrics",
		h.reportType,
	)
}

func (h *helper) getVmCpuMetrics() (interface{}, error) {
	if h.reportType == "rank" {
		return cubecos.GetVmCpuRank(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get vm cpu metrics",
		h.reportType,
	)
}
