package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getMemoryUsageMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return h.getMemoryUsageSummary()
	case "history":
		return nil, fmt.Errorf("history is not supported yet for cpu metrics")
	case "rank":
		return h.getMemoryUsageRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get memory metrics",
		h.viewType,
	)
}

func (h *helper) getMemoryUsageSummary() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetMemoryUsageSummaryOfHosts()
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for memory summary")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get memory summary",
		h.entityType,
	)
}

func (h *helper) getMemoryUsageRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetMemoryUsageRankOfHosts()
	case "vms":
		return cubecos.GetMemoryUsageRankOfVms(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get memory rank",
		h.entityType,
	)
}
