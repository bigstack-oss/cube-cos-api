package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
)

func (h *helper) getCpuUsage() (any, error) {
	switch h.viewType {
	case "summary":
		return h.getCpuUsageSummary()
	case "history":
		return h.getCpuUsageHistory()
	case "rank":
		return h.getCpuUsageRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get cpu metrics",
			h.viewType,
		)
	}
}

func (h *helper) getCpuUsageSummary() (any, error) {
	switch h.entityType {
	case "host":
		return cubecos.GetHostCpuSummary(h.entityId)
	case "hosts":
		return cubecos.GetHostsCpuSummary(h.genHostsCpuSummaryStmt())
	case "vm":
		return nil, metric.ErrVmNotSupportCpuSummary
	case "vms":
		return nil, metric.ErrVmNotSupportCpuSummary
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get cpu summary",
			h.entityType,
		)
	}
}

func (h *helper) getCpuUsageHistory() (any, error) {
	switch h.entityType {
	case "host":
		stmt := h.genHostCpuUsageHistoryStmt()
		return cubecos.GetHostCpuHistory(stmt)
	case "vm":
		return nil, metric.ErrVmNotSupportCpuHistory
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get cpu history",
			h.entityType,
		)
	}
}

func (h *helper) getCpuUsageRank() (any, error) {
	switch h.entityType {
	case "hosts":
		stmt := h.genHostCpuUsageRankStmt()
		return cubecos.GetHostsCpuUsageRank(stmt)
	case "vms":
		stmt := h.genVmCpuUsageRankStmt()
		return cubecos.GetVmsCpuUsageRank(stmt)
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get cpu rank",
			h.entityType,
		)
	}
}
