package cubecos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
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

func GetDataCenterFirmwareVersion() (string, error) {
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

func GetDataCenterFixpackVersion() (string, error) {
	return "", nil
}

func GetDataCenterFixpackUpdatedAt() (string, error) {
	return "", nil
}

func GetFirmwareLastUpdatedAt() (string, error) {
	partition, err := GetActivePartition()
	if err != nil {
		return "", err
	}

	timestamp, err := ReadPartitionInstallTimestamp(partition)
	if err != nil {
		return "", fmt.Errorf("failed to read last update time from partition %s: %v", partition, err)
	}

	t := ostime.Unix(timestamp, 0)
	return time.LocalRFC3339(t), nil
}

func GetActivePartition() (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(10))
	defer cancel()

	out, err := exec.CommandContext(ctx, "/usr/sbin/grub-get-default").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"failed to get active partition(%v %s)",
			err, string(out),
		)
	}

	if len(out) == 0 {
		return "", fmt.Errorf("active partition is empty")
	}

	activeNum := strings.TrimSpace(string(out))
	return activeNum, nil
}

func ReadPartitionInstallTimestamp(partition string) (int64, error) {
	path := fmt.Sprintf("/boot/grub2/info%s", partition)
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read partition date from %s(%v)", path, err)
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("partition data is empty in %s", path)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("no lines found in partition data")
	}

	for _, line := range lines {
		if !strings.HasPrefix(line, "install_date") {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid install date format: %s", line)
		}

		raw := strings.TrimSpace(parts[1])
		timestamp, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse install date(%v)", err)
		}

		return timestamp, nil
	}

	return 0, fmt.Errorf(
		"install date not found in partition data",
	)
}
