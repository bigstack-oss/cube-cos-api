package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getNetworkIngressTraffic() (any, error) {
	switch h.viewType {
	case "rank":
		return h.getNetworkIngressRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get network traffic in metrics",
			h.viewType,
		)
	}
}

func (h *helper) getNetworkIngressRank() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetHostsNetworkIngressRank()
	case "vms":
		return cubecos.GetVmsNetworkIngressRank(h.genVmsNetworkIngressRankStmt())
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get network traffic in rank",
			h.entityType,
		)
	}
}

func (h *helper) getNetworkEgressTraffic() (any, error) {
	switch h.viewType {
	case "rank":
		return h.getNetworkEgressRank()
	default:
		return nil, fmt.Errorf(
			"invalid view type(%s) to get network traffic out metrics",
			h.viewType,
		)
	}
}

func (h *helper) getNetworkEgressRank() (any, error) {
	switch h.entityType {
	case "hosts":
		return cubecos.GetHostsNetworkEgressRank()
	case "vms":
		return cubecos.GetVmsNetworkEgressRank(h.genVmsNetworkEgressRankStmt())
	default:
		return nil, fmt.Errorf(
			"invalid entity type(%s) to get network traffic out rank",
			h.entityType,
		)
	}
}
