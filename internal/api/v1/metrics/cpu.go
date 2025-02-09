package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getCpuUsageMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return h.getCpuUsageSummary()
	case "history":
		return nil, fmt.Errorf("history is not supported yet for cpu metrics")
	case "rank":
		return h.getCpuUsageRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get cpu metrics",
		h.viewType,
	)
}

func (h *helper) getCpuUsageSummary() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetCpuSummaryOfHosts()
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for cpu summary")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get cpu summary",
		h.entityType,
	)
}
func (h *helper) getCpuUsageRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetCpuUsageRankOfHosts(h.rank.head)
	case "vms":
		return cubecos.GetCpuUsageRankOfVms(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get cpu rank",
		h.entityType,
	)
}
