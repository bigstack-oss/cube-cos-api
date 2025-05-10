package datacenters

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
)

func getLocalDataCenter() datacenters.DataCenter {
	return datacenters.DataCenter{
		Type:        getDataCenterType(),
		Roles:       getDataCenterAllowRoles(),
		Name:        base.DataCenterName,
		Version:     base.DataCenterVersion,
		VirtualIp:   base.DataCenterVip,
		IsLocal:     true,
		IsHaEnabled: base.IsHaEnabled,
		UtcTimeZone: v1.LocalTimeZone,
		Additional: datacenters.Additional{
			HelpUrl:           base.DataCenterHelpUrl,
			NodeLicenseStatus: getNodeLicenseStatus(),
		},
	}
}

func getDataCenterType() string {
	for _, node := range nodes.List() {
		if datacenters.IsCloudRole(node.Role) {
			return "cloud"
		}

		if datacenters.IsEdgeRole(node.Role) {
			return "edge"
		}
	}

	return "unknown"
}

func getDataCenterAllowRoles() []string {
	switch getDataCenterType() {
	case "edge":
		return datacenters.GetEdgeRoles()
	case "cloud":
		return datacenters.GetCloudRoles()
	}

	return []string{}
}
