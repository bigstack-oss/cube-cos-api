package service

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/error"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

func getNodesByService(svc *registry.Service, role string) []definition.Node {
	nodes := []definition.Node{}

	for _, node := range svc.Nodes {
		if isCurrentNode(node) {
			continue
		}

		nodes = append(
			nodes,
			genNodeInfo(node, role),
		)
	}

	return nodes
}

func genNodeInfo(node *registry.Node, role string) definition.Node {
	return definition.Node{
		Role:     role,
		ID:       definition.HostID,
		Hostname: definition.Hostname,
		Address:  node.Address,
	}
}

func isCurrentNode(node *registry.Node) bool {
	return node.Address == definition.AdvertiseAddr
}

func GetNodesByRole(roleName string) ([]definition.Node, error) {
	svcs, err := registry.GetService(roleName)
	if err != nil {
		log.Errorf("failed to get service %s (%s)", roleName, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		return nil, cuberr.ServiceNotFound
	}

	nodes := []definition.Node{}
	for _, svc := range svcs {
		nodes = append(
			nodes,
			getNodesByService(svc, roleName)...,
		)
	}

	return nodes, nil
}
