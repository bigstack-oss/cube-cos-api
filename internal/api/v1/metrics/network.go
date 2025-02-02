package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getNetworkMetrics() (interface{}, error) {
	switch h.resourceType {
	case "hosts":
		return h.getHostNetworkMetrics()
	case "vms":
		return h.getVmNetworkMetrics()
	}

	return nil, fmt.Errorf(
		"invalid resource type(%s) to get network metrics",
		h.resourceType,
	)
}

func (h *helper) getHostNetworkMetrics() (interface{}, error) {
	switch h.reportType {
	case "rank":
		return h.getHostNetworkRank()
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get host network metrics",
		h.reportType,
	)
}

func (h *helper) getHostNetworkRank() (interface{}, error) {
	switch h.metricType {
	case "ingress":
		return cubecos.GetHostNetworkIngressRank()
	case "egress":
		return cubecos.GetHostNetworkEgressRank()
	}

	return nil, fmt.Errorf(
		"invalid metric type(%s) to get host network rank",
		h.metricType,
	)
}

func (h *helper) getVmNetworkMetrics() (interface{}, error) {
	switch h.reportType {
	case "rank":
		return h.getVmNetworkRank()
	}

	return nil, fmt.Errorf(
		"invalid report type(%s) to get vm network metrics",
		h.reportType,
	)
}

func (h *helper) getVmNetworkRank() (interface{}, error) {
	switch h.metricType {
	case "ingress":
		return cubecos.GetVmsNetworkIngressRank(h.rank.head)
	case "egress":
		return cubecos.GetVmsNetworkEgressRank(h.rank.head)
	}

	return nil, fmt.Errorf(
		"invalid metric type(%s) to get vms net rank metrics",
		h.metricType,
	)
}
