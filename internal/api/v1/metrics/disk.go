package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getDiskBandwidth() (any, error) {
	switch h.viewType {
	case "history":
		return h.getDiskBandwidthHistory()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get disk bandwidth metrics",
		h.viewType,
	)
}

func (h *helper) getDiskBandwidthHistory() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetHostsDiskBandwidthHistory(
			h.genHostsDiskReadBandwidthStmt(),
			h.genHostsDiskWriteBandwidthStmt(),
		)
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get disk bandwidth history",
			h.entityType,
		)
	}
}

func (h *helper) getDiskUsage() (any, error) {
	switch h.viewType {
	case "rank":
		return h.getDiskUsageRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get disk usage metrics",
			h.viewType,
		)
	}
}

func (h *helper) getDiskUsageRank() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetHostsDiskUsageRank(h.genHostsStorageUsageRankStmt())
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get disk usage rank",
			h.entityType,
		)
	}
}

func (h *helper) getDiskIops() (any, error) {
	switch h.viewType {
	case "history":
		return h.getDiskIopsHistory()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get disk iops metrics",
			h.viewType,
		)
	}
}

func (h *helper) getDiskIopsHistory() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetHostsDiskIopsHistory(
			h.genHostsStorageReadIopsStmt(),
			h.genHostsStorageWriteIopsStmt(),
		)
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get disk iops",
			h.entityType,
		)
	}
}

func (h *helper) getDiskReadIops() (any, error) {
	switch h.viewType {
	case "rank":
		return h.getDiskReadIopsRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get disk read iops metrics",
			h.viewType,
		)
	}
}

func (h *helper) getDiskReadIopsRank() (any, error) {
	switch h.entityType {
	case "vms":
		return cubecos.GetVmsDiskReadIopsRank(h.genVmsStorageIopsReadRankStmt())
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get disk read iops rank",
			h.entityType,
		)
	}
}

func (h *helper) getDiskWriteIops() (any, error) {
	switch h.viewType {
	case "rank":
		return h.getDiskWriteIopsRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get disk write iops metrics",
			h.viewType,
		)
	}
}

func (h *helper) getDiskWriteIopsRank() (any, error) {
	switch h.entityType {
	case "vms":
		return cubecos.GetVmsDiskWriteIopsRank(h.genVmsStorageIopsWriteRankStmt())
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get disk write iops rank",
			h.entityType,
		)
	}
}

func (h *helper) getDiskLatency() (any, error) {
	switch h.viewType {
	case "history":
		return h.getDiskLatencyHistory()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get disk latency metrics",
			h.viewType,
		)
	}
}

func (h *helper) getDiskLatencyHistory() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GeHostsDiskLatencyHistory(
			h.genHostsStorageReadLatencyStmt(),
			h.genHostsStorageWriteLatencyStmt(),
		)
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get disk latency",
			h.entityType,
		)
	}
}
