package service

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

func parseNodes(svc *registry.Service) []definition.Node {
	nodes := []definition.Node{}

	for _, node := range svc.Nodes {
		if isCurrentNode(node) {
			continue
		}

		nodes = append(nodes, newNode(node))
	}

	return nodes
}

func newNode(node *registry.Node) definition.Node {
	return definition.Node{
		Role:     node.Metadata["role"],
		Id:       definition.HostID,
		Hostname: definition.Hostname,
		Address:  node.Address,
	}
}

func isCurrentNode(node *registry.Node) bool {
	return node.Address == definition.AdvertiseAddr
}

func GetNodesByRole(roleName string) ([]definition.Node, error) {
	svcs, err := registry.GetService(definition.DataCenterName)
	if err != nil {
		log.Errorf("failed to get service %s (%s)", roleName, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		return nil, cuberr.ServiceNotFound
	}

	nodes := []definition.Node{}
	for _, svc := range svcs {
		roleNodes := parseNodes(svc)
		if len(roleNodes) == 0 {
			continue
		}

		nodes = append(nodes, roleNodes...)
	}

	return nodes, nil
}
