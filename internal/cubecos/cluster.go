package cubecos

import (
	"os"
	"os/exec"
	"strconv"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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

func IsClusterSetReady() bool {
	_, err := exec.Command("hex_sdk", "cube_cluster_ready").Output()
	if err != nil {
		return IsExpectedEmptyStdOut(err)
	}

	return true
}

func GetDataCenterName() (string, error) {
	if !definition.IsHaEnabled {
		return os.Hostname()
	}

	return ReadHexTuning(CubeSysController)
}

func GetDataCenterSummary() (*Summary, error) {
	resourceMetrics, err := GetResourceMetrics()
	if err != nil {
		return nil, err
	}

	roleOverview, err := GetRoleOverview()
	if err != nil {
		return nil, err
	}

	vmStatusOverview, err := GetVmStatusOverview()
	if err != nil {
		return nil, err
	}

	return &Summary{
		Metrics: *resourceMetrics,
		Role:    *roleOverview,
		Vm:      *vmStatusOverview,
	}, nil
}
