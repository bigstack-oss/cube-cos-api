package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getNetworkTrafficInMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for network traffic in metrics")
	case "history":
		return nil, fmt.Errorf("history is not supported yet for network traffic in metrics")
	case "rank":
		return h.getNetworkTrafficInRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get network traffic in metrics",
		h.viewType,
	)
}

func (h *helper) getNetworkTrafficInRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetNetworkTrafficInRankOfHosts()
	case "vms":
		return cubecos.GetNetworkTrafficInRankOfVms(h.genVmNetworkIngressRankStmt())
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get network traffic in rank",
		h.entityType,
	)
}

func (h *helper) getNetworkTrafficOutMetrics() (interface{}, error) {
	switch h.viewType {
	case "summary":
		return nil, fmt.Errorf("summary is not supported yet for network traffic out metrics")
	case "history":
		return nil, fmt.Errorf("history is not supported yet for network traffic out metrics")
	case "rank":
		return h.getNetworkTrafficOutRank()
	}

	return nil, fmt.Errorf(
		"invalid view type(%s) to get network traffic out metrics",
		h.viewType,
	)
}

func (h *helper) getNetworkTrafficOutRank() (interface{}, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetNetworkTrafficOutRankOfHosts()
	case "vms":
		return cubecos.GetNetworkTrafficOutRankOfVms(h.genVmNetworkEgressRankStmt())
	}

	return nil, fmt.Errorf(
		"invalid entity type(%s) to get network traffic out rank",
		h.entityType,
	)
}
