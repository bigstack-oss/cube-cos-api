package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getMemoryMetrics() (interface{}, error) {
	switch h.resourceType {
	case "hosts":
		return h.getHostMemoryMetrics()
	case "vms":
		return h.getVmMemoryMetrics()
	}

	return nil, fmt.Errorf(
		"invalid resource type(%s) to get memory metrics",
		h.resourceType,
	)
}

func (h *helper) getHostMemoryMetrics() (interface{}, error) {
	switch h.reportType {
	case "summary":
		return cubecos.GetHostMemorySummary()
	case "rank":
		return cubecos.GetHostMemoryRank()
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get host memory metrics",
		h.reportType,
	)
}

func (h *helper) getVmMemoryMetrics() (interface{}, error) {
	if h.reportType == "rank" {
		return cubecos.GetVmMemoryRank(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get vm memory metrics",
		h.reportType,
	)
}
