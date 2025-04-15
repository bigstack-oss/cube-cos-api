package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getMemoryUsageMetrics() (any, error) {
	switch h.viewType {
	case "summary":
		return h.getMemoryUsageSummary()
	case "history":
		return h.getMemoryHistory()
	case "rank":
		return h.getMemoryUsageRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get memory metrics",
		h.viewType,
	)
}

func (h *helper) getMemoryUsageSummary() (any, error) {
	switch h.entityType {
	case "host":
		return cubecos.GetMemoryUsageSummaryOfHost(h.entityId)
	case "hosts":
		return cubecos.GetMemoryUsageSummaryOfHosts()
	case "vm":
		return nil, fmt.Errorf("vm is not supported yet for memory summary")
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for memory summary")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get memory summary",
		h.entityType,
	)
}

func (h *helper) getMemoryHistory() (any, error) {
	switch h.entityType {
	case "host":
		stmt := h.genHostMemorySizeHistoryStmt()
		return cubecos.GetMemorySizeHistoryOfHost(stmt)
	case "vm":
		return nil, fmt.Errorf("vm is not supported yet for memory history")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get memory history",
		h.entityType,
	)
}

func (h *helper) getMemoryUsageRank() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetMemoryUsageRankOfHosts(h.genHostMemoryUsageRankStmt())
	case "vms":
		return cubecos.GetMemoryUsageRankOfVms(h.genVmMemoryRankStmt())
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get memory rank",
		h.entityType,
	)
}
