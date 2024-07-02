package tuning

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func listNodesBySelector(nodes []*definition.Node, selector definition.Selector) []*definition.Node {
	if !selector.Enabled {
		return nodes
	}

	filteredNodes := []*definition.Node{}
	for _, node := range nodes {
		for key, value := range selector.Labels {
			if node.Labels[key] == value {
				filteredNodes = append(filteredNodes, node)
				break
			}
		}
	}

	return filteredNodes
}

func selectRolesUsingActivityAndLabels(tuningSpec *definition.TuningSpec) []*definition.Role {
	for i, role := range tuningSpec.Roles {
		tuningSpec.Roles[i].Nodes = listNodesBySelector(role.Nodes, tuningSpec.Selector)
	}

	roles := []*definition.Role{}
	for _, role := range tuningSpec.Roles {
		if !role.IsNodeEmpty() {
			roles = append(roles, role)
		}
	}

	return roles
}
