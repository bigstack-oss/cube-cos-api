package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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
		return cubecos.GetCpuSummaryOfHost(h.entityId)
	case "hosts":
		return cubecos.GetCpuSummaryOfHosts(h.genHostCpuUsageStmt())
	case "vm":
		return nil, fmt.Errorf("vm is not supported yet for cpu summary")
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for cpu summary")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get cpu summary",
		h.entityType,
	)
}

func (h *helper) getCpuUsageHistory() (any, error) {
	switch h.entityType {
	case "host":
		stmt := h.genHostCpuUsageHistoryStmt()
		return cubecos.GetCpuHistoryOfHost(stmt)
	case "vm":
		return nil, fmt.Errorf("vm is not supported yet for cpu history")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get cpu history",
		h.entityType,
	)
}

func (h *helper) getCpuUsageRank() (any, error) {
	switch h.entityType {
	case "hosts":
		stmt := h.genHostCpuUsageRankStmt()
		return cubecos.GetCpuUsageRankOfHosts(stmt)
	case "vms":
		stmt := h.genVmCpuUsageRankStmt()
		return cubecos.GetCpuUsageRankOfVms(stmt)
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get cpu rank",
		h.entityType,
	)
}
