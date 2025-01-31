package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getStorageMetrics() (interface{}, error) {
	switch h.resourceType {
	case "hosts":
		return h.getHostStorageMetrics()
	case "vms":
		return h.getVmStorageMetrics()
	}

	return nil, fmt.Errorf(
		"invalid resource type(%s) to get storage metrics",
		h.resourceType,
	)
}

func (h *helper) getHostStorageMetrics() (interface{}, error) {
	switch h.reportType {
	case "timeSeries":
		return h.getHostStorageTimeSeries()
	case "rank":
		return h.getHostStorageRank()
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get host storage metrics",
		h.reportType,
	)
}

func (h *helper) getHostStorageTimeSeries() (interface{}, error) {
	switch h.metricType {
	case "bandwidth":
		return cubecos.GetHostStorageBandwidthSeries(h.Period)
	case "iops":
		return cubecos.GetHostStorageIopsSeries(h.Period)
	case "latency":
		return cubecos.GetHostStorageLatencySeries(h.Period)
	}

	return nil, fmt.Errorf(
		"invalid metric type(%s) to get host storage time series",
		h.metricType,
	)
}

func (h *helper) getHostStorageRank() (interface{}, error) {
	if h.metricType == "usage" {
		return cubecos.GetHostStorageUsageRank(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid metric type(%s) to get host storage rank",
		h.metricType,
	)
}

func (h *helper) getVmStorageMetrics() (interface{}, error) {
	if h.reportType == "rank" {
		return h.getVmStorageOptsRank()
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get vm storage metrics",
		h.reportType,
	)
}

func (h *helper) getVmStorageOptsRank() (interface{}, error) {
	switch h.metricType {
	case "iopsRead":
		return cubecos.GetVmsStorageIopsReadRank(h.rank.head)
	case "iopsWrite":
		return cubecos.GetVmsStorageIopsWriteRank(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid metric type(%s) to get vm storage rank",
		h.metricType,
	)
}
