package datacenter

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
)

func IsCloudType() bool {
	for _, node := range nodes.List() {
		if nodes.IsCloudRole(node.Role) {
			return true
		}
	}

	return false
}

func GetType() string {
	for _, node := range nodes.List() {
		if nodes.IsCloudRole(node.Role) {
			return base.Cloud
		}

		if nodes.IsEdgeRole(node.Role) {
			return base.Edge
		}
	}

	return "unknown"
}

func GetAllowRoles() []string {
	switch GetType() {
	case base.Edge:
		return nodes.GetEdgeRoles()
	case base.Cloud:
		return nodes.GetCloudRoles()
	}

	return []string{}
}
