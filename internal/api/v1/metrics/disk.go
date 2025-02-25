package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getDiskBandwidthMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for disk bandwidth metrics")
	case "history":
		return h.getDiskBandwidthHistory()
	case "rank":
		return nil, fmt.Errorf("rank is not supported yet for disk bandwidth metrics")
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk bandwidth metrics",
		h.viewType,
	)
}

func (h *helper) getDiskBandwidthHistory() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetDiskStorageBandwidthHistory(
			h.genHostStorageReadBandwidthStmt(),
			h.genHostStorageWriteBandwidthStmt(),
		)
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for disk bandwidth history")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get disk bandwidth history",
		h.entityType,
	)
}

func (h *helper) getDiskUsageMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for disk usage metrics")
	case "history":
		return nil, fmt.Errorf("history is not supported yet for disk usage metrics")
	case "rank":
		return h.getDiskUsageRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk usage metrics",
		h.viewType,
	)
}

func (h *helper) getDiskUsageRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetDiskUsageRankOfHosts(h.genHostStorageUsageRankStmt())
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for disk usage rank")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get disk usage rank",
		h.entityType,
	)
}

func (h *helper) getDiskIopsMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for disk iops metrics")
	case "history":
		return h.getDiskIopsHistory()
	case "rank":
		return nil, fmt.Errorf("rank is not supported yet for disk iops metrics")
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk iops metrics",
		h.viewType,
	)
}

func (h *helper) getDiskIopsHistory() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetDiskIopsHistoryOfHosts(
			h.genHostStorageReadIopsStmt(),
			h.genHostStorageWriteIopsStmt(),
		)
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for disk iops")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get disk iops",
		h.entityType,
	)
}

func (h *helper) getDiskReadIopsMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for disk read iops metrics")
	case "history":
		return nil, fmt.Errorf("history is not supported yet for disk read iops metrics")
	case "rank":
		return h.getDiskReadIopsRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk read iops metrics",
		h.viewType,
	)
}

func (h *helper) getDiskReadIopsRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return nil, fmt.Errorf("hosts is not supported yet for disk read iops rank")
	case "vms":
		return cubecos.GetDiskReadIopsRankOfVms(h.genVmStorageIopsReadRankStmt())
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get disk read iops rank",
		h.entityType,
	)
}

func (h *helper) getDiskWriteIopsMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for disk write iops metrics")
	case "history":
		return nil, fmt.Errorf("history is not supported yet for disk write iops metrics")
	case "rank":
		return h.getDiskWriteIopsRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk write iops metrics",
		h.viewType,
	)
}

func (h *helper) getDiskWriteIopsRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return nil, fmt.Errorf("hosts is not supported yet for disk write iops rank")
	case "vms":
		return cubecos.GetDiskWriteIopsRankOfVms(h.genVmStorageIopsWriteRankStmt())
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get disk write iops rank",
		h.entityType,
	)
}

func (h *helper) getDiskLatencyMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for disk latency metrics")
	case "history":
		return h.getDiskLatencyHistory()
	case "rank":
		return nil, fmt.Errorf("rank is not supported yet for disk latency metrics")
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk latency metrics",
		h.viewType,
	)
}

func (h *helper) getDiskLatencyHistory() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GeDiskLatencyHistoryOfHosts(
			h.genHostStorageReadLatencyStmt(),
			h.genHostStorageWriteLatencyStmt(),
		)
	case "vms":
		return nil, fmt.Errorf("vms is not supported yet for disk latency")
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get disk latency",
		h.entityType,
	)
}
