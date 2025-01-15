package cubecos

const ()

func GetControllerVirtualIp() (string, error) {
	value, err := ReadHexTuning(CubeSysControllerVip)
	if err != nil {
		return "", err
	}

	return value, nil
}
