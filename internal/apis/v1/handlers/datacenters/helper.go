package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

func getLocalDataCenter() base.DataCenter {
	return base.DataCenter{
		Type:        getDataCenterType(),
		Roles:       getDataCenterAllowRoles(),
		Name:        base.DataCenterName,
		Version:     base.DataCenterVersion,
		VirtualIp:   base.DataCenterVip,
		IsLocal:     true,
		IsHaEnabled: base.IsHaEnabled,
		UtcTimeZone: time.LocalZone,
		Additional: base.Additional{
			HelpUrl:           base.DataCenterHelpUrl,
			NodeLicenseStatus: getNodeLicenseStatus(),
		},
	}
}

func getDataCenterType() string {
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

func getDataCenterAllowRoles() []string {
	switch getDataCenterType() {
	case base.Edge:
		return nodes.GetEdgeRoles()
	case base.Cloud:
		return nodes.GetCloudRoles()
	}

	return []string{}
}
