package cubecos

import (
	"os"
	"strconv"
)

func IsHaEnabled() (bool, error) {
	strIsHaEnabled, err := ReadHexTuning(CubeSysHa)
	if err != nil {
		return false, err
	}

	isHaEnabled, err := strconv.ParseBool(strIsHaEnabled)
	if err != nil {
		return false, err
	}

	return isHaEnabled, nil
}

// M2 TODO: Check if the data center is local
func IsLocalDataCenter(dataCenter string) bool {
	return true
}

func GetDataCenterName(isHaEnabled bool) (string, error) {
	if !isHaEnabled {
		return os.Hostname()
	}

	return ReadHexTuning(CubeSysController)
}
