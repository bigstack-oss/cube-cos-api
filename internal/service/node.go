package service

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"go-micro.dev/v5/registry"
)

func parseNodes(svc *registry.Service) []v1.Node {
	nodes := []v1.Node{}

	for _, node := range svc.Nodes {
		if isLocalNode(node) {
			continue
		}

		nodes = append(nodes, newNode(node))
	}

	return nodes
}

func newNode(node *registry.Node) v1.Node {
	return v1.Node{
		Role:       node.Metadata["role"],
		Id:         v1.HostID,
		DataCenter: node.Metadata["dataCenter"],
		Protocol:   node.Metadata["protocol"],
		Ip:         node.Metadata["ip"],
		Hostname:   v1.Hostname,
		Token:      node.Metadata["token"],
		Address:    node.Address,
	}
}

func isLocalNode(node *registry.Node) bool {
	return node.Address == v1.AdvertiseAddr
}

func GetNodesByRole(roleName string) ([]v1.Node, error) {
	svcs, err := v1.GetRegisteredServices()
	if err != nil {
		return nil, err
	}

	nodes := []v1.Node{}
	for _, svc := range svcs {
		roleNodes := parseNodes(svc)
		setNodesIfRoleMatched(&nodes, roleNodes, roleName)
	}

	if len(nodes) <= 0 {
		return nil, fmt.Errorf("no nodes found for role %s", roleName)
	}

	return nodes, nil
}

func setNodesIfRoleMatched(nodes *[]v1.Node, roleNodes []v1.Node, roleName string) {
	for _, node := range roleNodes {
		if node.Role == roleName {
			*nodes = append(*nodes, node)
		}
	}
}
