package cubecos

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func GetSystemSeed() (string, error) {
	return GetTuningValue(CubeSysSeed)
}

func GetBoardSerial() (string, error) {
	data, err := os.ReadFile(base.BoardSerialPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), err
}

func GetDataCenterNumericVersion() (string, error) {
	return ReadSettingSys(SysProductVersion)
}

// M2 TODO: Check if the data center is local
func IsLocalDataCenter(dataCenter string) bool {
	return true
}

func IsHaEnabled() (bool, error) {
	strIsHaEnabled, err := GetTuningValue(CubeSysHa)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(strIsHaEnabled)
}

func IsDataCenterReady() bool {
	_, err := exec.Command("hex_sdk", "cube_cluster_ready").Output()
	if err != nil {
		return IsExpectedEmptyStdOut(err)
	}

	return true
}

func GetDataCenterName() (string, error) {
	if !base.IsHaEnabled {
		return os.Hostname()
	}

	return GetTuningValue(CubeSysController)
}

func GetDataCenterVersion() (string, error) {
	desc, err := ReadSettingSys(SysProductDescription)
	if err != nil {
		return "", err
	}

	version, err := ReadSettingSys(SysProductVersion)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", desc, version), nil
}
