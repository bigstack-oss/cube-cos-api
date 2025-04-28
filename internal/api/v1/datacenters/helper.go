package datacenters

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func getLocalDataCenter() v1.DataCenter {
	return v1.DataCenter{
		Type:        getDataCenterType(),
		Roles:       getDataCenterAllowRoles(),
		Name:        v1.DataCenterName,
		Version:     v1.DataCenterVersion,
		VirtualIp:   v1.DataCenterVip,
		IsLocal:     true,
		IsHaEnabled: v1.IsHaEnabled,
		UtcTimeZone: v1.LocalTimeZone,
		Additional: v1.Additional{
			HelpUrl:           v1.DataCenterHelpUrl,
			NodeLicenseStatus: getNodeLicenseStatus(),
		},
	}
}

func getDataCenterType() string {
	for _, node := range v1.ListNodes() {
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
