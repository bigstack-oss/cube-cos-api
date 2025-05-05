package cubecos

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

const ()

func GetControllerVirtualIp(mgmtNet string) (string, error) {
	if !v1.IsHaEnabled {
		return GetStandaloneVirtualIp(mgmtNet)
	}

	return GetClusterVirtualIp()
}

func GetStandaloneVirtualIp(mgmtNet string) (string, error) {
	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return GetTuningValue(netIfAddrMgmtIp)
}

func GetClusterVirtualIp() (string, error) {
	switch v1.CurrentRole {
	case v1.RoleControl, v1.RoleControlConverged, v1.RoleModerator:
		return GetTuningValue(CubeSysControllerVip)
	case v1.RoleCompute, v1.RoleStorage, v1.RoleEdgeCore:
		return GetTuningValue(CubeSysControllerIp)
	}

	return "", fmt.Errorf(
		"unsupported role for reading cluster virtual ip: %s",
		v1.CurrentRole,
	)
}

func GetManagementNet() (string, error) {
	return GetTuningValue(CubeSysManagementNetwork)
}

func GetManagementIp(mgmtNet string) (string, error) {
	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return GetTuningValue(netIfAddrMgmtIp)
}

func GetStorageNet() (string, error) {
	return GetTuningValue(CubeSysStorageNetwork)
}

func GetStorageIp(storageNet string) (string, error) {
	if storageNet == "" {
		return "", fmt.Errorf("storage network is empty")
	}

	netIfAddrStorageIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, storageNet)
	return GetTuningValue(netIfAddrStorageIp)
}
