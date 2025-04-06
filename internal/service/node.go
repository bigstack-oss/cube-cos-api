package service

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"go-micro.dev/v5/registry"
)

func parseNodes(svc *registry.Service) []definition.Node {
	nodes := []definition.Node{}

	for _, node := range svc.Nodes {
		if isLocalNode(node) {
			continue
		}

		nodes = append(nodes, newNode(node))
	}

	return nodes
}

func newNode(node *registry.Node) definition.Node {
	return definition.Node{
		Role:       node.Metadata["role"],
		Id:         definition.HostID,
		DataCenter: node.Metadata["dataCenter"],
		Protocol:   node.Metadata["protocol"],
		Ip:         node.Metadata["ip"],
		Hostname:   definition.Hostname,
		Token:      node.Metadata["token"],
		Address:    node.Address,
	}
}

func isLocalNode(node *registry.Node) bool {
	return node.Address == definition.AdvertiseAddr
}

func GetNodesByRole(roleName string) ([]definition.Node, error) {
	svcs, err := definition.GetRegisteredServices()
	if err != nil {
		return nil, err
	}

	nodes := []definition.Node{}
	for _, svc := range svcs {
		roleNodes := parseNodes(svc)
		setNodesIfRoleMatched(&nodes, roleNodes, roleName)
	}

	if len(nodes) <= 0 {
		return nil, fmt.Errorf("no nodes found for role %s", roleName)
	}

	return nodes, nil
}

func setNodesIfRoleMatched(nodes *[]definition.Node, roleNodes []definition.Node, roleName string) {
	for _, node := range roleNodes {
		if node.Role == roleName {
			*nodes = append(*nodes, node)
		}
	}
}
