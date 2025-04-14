package cubecos

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

const ()

func GetControllerVirtualIp(mgmtNet string) (string, error) {
	if definition.IsHaEnabled {
		return GetTuningValue(CubeSysControllerVip)
	}

	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return GetTuningValue(netIfAddrMgmtIp)
}

func GetMgmtNet() (string, error) {
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

func GetStorageIp(stroageNet string) (string, error) {
	if stroageNet == "" {
		return "", fmt.Errorf("storage network is empty")
	}

	netIfAddrStorageIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, stroageNet)
	return GetTuningValue(netIfAddrStorageIp)
}
