package nodes

import (
	"go-micro.dev/v5/registry"
)

func HostnameMap() (map[string]Node, error) {
	svcs, err := GetDiscoveredServices()
	if err != nil {
		return nil, err
	}

	nodeMap := map[string]Node{}
	for _, svc := range svcs {
		nodes := parseNodes(svc)
		for _, node := range nodes {
			nodeMap[node.Hostname] = node
		}
	}

	return nodeMap, nil
}

func parseNodes(svc *registry.Service) []Node {
	nodes := []Node{}

	for _, node := range svc.Nodes {
		if IsLocal(node.Metadata["hostname"]) {
			continue
		}

		nodes = append(nodes, New(node))
	}

	return nodes
}

func convertToHosts(nodes []Node) []Host {
	hosts := []Host{}
	for _, node := range nodes {
		hosts = append(
			hosts,
			Host{
				Name: node.Hostname,
				Ip:   node.Ip,
			},
		)
	}

	return hosts
}

func parseNodesByRole(svc *registry.Service, role string) []Node {
	nodes := []Node{}
	for _, node := range svc.Nodes {
		if node.Metadata["role"] != role {
			continue
		}

		nodes = append(nodes, New(node))
	}

	return nodes
}
