package service

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"go-micro.dev/v5/registry"
)

func parseNodes(svc *registry.Service) []nodes.Node {
	nodes := []nodes.Node{}

	for _, node := range svc.Nodes {
		if isLocalNode(node) {
			continue
		}

		nodes = append(nodes, new(node))
	}

	return nodes
}

func new(n *registry.Node) nodes.Node {
	return nodes.Node{
		Role:       n.Metadata["role"],
		Id:         base.HostID,
		DataCenter: n.Metadata["dataCenter"],
		Protocol:   n.Metadata["protocol"],
		Ip:         n.Metadata["ip"],
		Hostname:   base.Hostname,
		Address:    n.Address,
	}
}

func isLocalNode(node *registry.Node) bool {
	return node.Address == base.AdvertiseAddr
}

func GetNodesByRole(roleName string) ([]nodes.Node, error) {
	svcs, err := nodes.GetDiscoveredServices()
	if err != nil {
		return nil, err
	}

	nodes := []nodes.Node{}
	for _, svc := range svcs {
		roleNodes := parseNodes(svc)
		setNodesIfRoleMatched(&nodes, roleNodes, roleName)
	}

	if len(nodes) <= 0 {
		return nil, fmt.Errorf("no nodes found for role %s", roleName)
	}

	return nodes, nil
}

func setNodesIfRoleMatched(nodes *[]nodes.Node, roleNodes []nodes.Node, roleName string) {
	for _, node := range roleNodes {
		if node.Role == roleName {
			*nodes = append(*nodes, node)
		}
	}
}
