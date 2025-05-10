package datacenters

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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
		UtcTimeZone: v1.LocalTimeZone,
		Additional: base.Additional{
			HelpUrl:           base.DataCenterHelpUrl,
			NodeLicenseStatus: getNodeLicenseStatus(),
		},
	}
}

func getDataCenterType() string {
	for _, node := range nodes.List() {
		if nodes.IsCloudRole(node.Role) {
			return "cloud"
		}

		if nodes.IsEdgeRole(node.Role) {
			return "edge"
		}
	}

	return "unknown"
}

func getDataCenterAllowRoles() []string {
	switch getDataCenterType() {
	case "edge":
		return nodes.GetEdgeRoles()
	case "cloud":
		return nodes.GetCloudRoles()
	}

	return []string{}
}
