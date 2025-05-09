package datacenters

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
)

func getLocalDataCenter() v1.DataCenter {
	return v1.DataCenter{
		Type:        getDataCenterType(),
		Roles:       getDataCenterAllowRoles(),
		Name:        base.DataCenterName,
		Version:     base.DataCenterVersion,
		VirtualIp:   base.DataCenterVip,
		IsLocal:     true,
		IsHaEnabled: base.IsHaEnabled,
		UtcTimeZone: v1.LocalTimeZone,
		Additional: v1.Additional{
			HelpUrl:           base.DataCenterHelpUrl,
			NodeLicenseStatus: getNodeLicenseStatus(),
		},
	}
}

func getDataCenterType() string {
	for _, node := range nodes.List() {
		if v1.IsCloudRole(node.Role) {
			return "cloud"
		}

		if v1.IsEdgeRole(node.Role) {
			return "edge"
		}
	}

	return "unknown"
}

func getDataCenterAllowRoles() []string {
	switch getDataCenterType() {
	case "edge":
		return v1.GetEdgeRoles()
	case "cloud":
		return v1.GetCloudRoles()
	}

	return []string{}
}
