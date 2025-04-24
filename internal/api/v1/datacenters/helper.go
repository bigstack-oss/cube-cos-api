package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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
	nodes, err := cubecos.ListNodes()
	if err != nil {
		return "unknown"
	}

	dataCenterTypes := map[string]int{}
	for _, node := range nodes {
		switch node.Role {
		case v1.RoleControlConverged, v1.RoleControl, v1.RoleCompute, v1.RoleStorage:
			dataCenterTypes["cloud"]++
		case v1.RoleModerator, v1.RoleEdgeCore:
			dataCenterTypes["edge"]++
		}
	}

	if dataCenterTypes["edge"] > 0 {
		return "edge"
	}

	if dataCenterTypes["cloud"] > 0 {
		return "cloud"
	}

	return "unknown"
}

func getDataCenterAllowRoles() []string {
	switch getDataCenterType() {
	case "edge":
		return []string{
			v1.RoleModerator,
			v1.RoleEdgeCore,
		}
	case "cloud":
		return []string{
			v1.RoleControlConverged,
			v1.RoleControl,
			v1.RoleCompute,
			v1.RoleStorage,
		}
	}

	return []string{}
}
