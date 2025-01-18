package cubecos

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

const ()

func GetControllerVirtualIp() (string, error) {
	if definition.IsHaEnabled {
		return ReadHexTuning(CubeSysControllerVip)
	}

	return ReadHexTuning(NetIfAddrEth0)
}
