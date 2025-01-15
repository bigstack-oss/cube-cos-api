package cubecos

import "strconv"

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
