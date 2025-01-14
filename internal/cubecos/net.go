package cubecos

const (
	cubeSysControllerVip = "cubesys.control.vip"
)

func GetControllerVirtualIp() (string, error) {
	value, err := HexTuningRead(cubeSysControllerVip)
	if err != nil {
		return "", err
	}

	return value, nil
}
