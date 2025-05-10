package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getMemoryUsage() (any, error) {
	switch h.viewType {
	case "summary":
		return h.getMemoryUsageSummary()
	case "history":
		return h.getMemoryUsageHistory()
	case "rank":
		return h.getMemoryUsageRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get memory metrics",
			h.viewType,
		)
	}
}

func (h *helper) getMemoryUsageSummary() (any, error) {
	switch h.entityType {
	case "host":
		return cubecos.GetHostMemoryUsageSummary(h.entityId)
	case "hosts":
		return cubecos.GetHostsMemoryUsageSummary()
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get memory summary",
			h.entityType,
		)
	}
}

func (h *helper) getMemoryUsageHistory() (any, error) {
	switch h.entityType {
	case "host":
		stmt := h.genHostMemorySizeHistoryStmt()
		return cubecos.GetHostMemorySizeHistory(stmt)
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get memory history",
			h.entityType,
		)
	}
}

func (h *helper) getMemoryUsageRank() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetHostsMemoryUsageRank(h.genHostsMemoryUsageRankStmt())
	case "vms":
		return cubecos.GetVmsMemoryUsageRank(h.genVmsMemoryRankStmt())
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get memory rank",
			h.entityType,
		)
	}
}
