package cubecos

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

const ()

func GetControllerVirtualIp(mgmtNet string) (string, error) {
	if definition.IsHaEnabled {
		return ReadTuning(CubeSysControllerVip)
	}

	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return ReadTuning(netIfAddrMgmtIp)
}

func GetMgmtNet() (string, error) {
	return ReadTuning(CubeSysManagementNetwork)
}

func GetManagementIp(mgmtNet string) (string, error) {
	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return ReadTuning(netIfAddrMgmtIp)
}
