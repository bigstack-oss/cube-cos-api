package cubecos

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

const ()

func GetControllerVirtualIp(mgmtNet string) (string, error) {
	if v1.IsHaEnabled {
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

func GetStorageIp(storageNet string) (string, error) {
	if storageNet == "" {
		return "", fmt.Errorf("storage network is empty")
	}

	netIfAddrStorageIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, storageNet)
	return GetTuningValue(netIfAddrStorageIp)
}
