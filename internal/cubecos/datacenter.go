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
	log "go-micro.dev/v5/logger"
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

func GetActiveFirmwareVersion() (string, error) {
	desc, err := ReadSettingSys(SysProductDescription)
	if err != nil {
		return "", err
	}

	version, err := ReadSettingSys(SysProductVersion)
	if err != nil {
		return "", err
	}

	date, err := ReadSettingSys(SysBuildLabel)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s %s", desc, version, date), nil
}

func GetInactiveFirmwareVersion() (string, error) {
	active, err := GetActivePartition()
	if err != nil {
		return "", err
	}

	inactive := ""
	if active == "1" {
		inactive = "2"
	} else {
		inactive = "1"
	}

	return ReadPartitionFirmwareVersion(inactive)
}

func GetFixpackVersion() (string, error) {
	fixpack, err := GetDataCenterLastInstalledFixpack()
	if err != nil {
		return "", err
	}

	version := fixpack[1]
	return version, nil
}

func GetFixpackUpdatedAt() (string, error) {
	fixpack, err := GetDataCenterLastInstalledFixpack()
	if err != nil {
		return "", err
	}

	updatedAt := fixpack[0]
	t, err := ostime.Parse(time.FormatFixpack, updatedAt)
	if err != nil {
		return "", fmt.Errorf("failed to parse fixpack updatedAt(%s %v)", updatedAt, err)
	}

	return time.LocalRFC3339(t), nil
}

func GetDataCenterLastInstalledFixpack() ([]string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	out, err := exec.CommandContext(ctx, "hex_config", "fixpack_get_history").Output()
	if err != nil {
		err = fmt.Errorf("failed to get data center fixpack version(%v %s)", err, string(out))
		log.Errorf("datacenter: %v", err)
		return nil, err
	}

	if len(out) == 0 {
		err := fmt.Errorf("data center fixpack version is empty")
		log.Errorf("datacenter: %v", err)
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	line := lines[len(lines)-2]
	segments := strings.Split(line, "|")
	if len(segments) < 7 {
		err := fmt.Errorf("invalid fixpack version format: %s", line)
		log.Errorf("datacenter: %v", err)
		return nil, err
	}

	return segments, nil
}

func GetActiveFirmwaretUpdatedAt() (string, error) {
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

func ReadPartitionFirmwareVersion(partition string) (string, error) {
	path := fmt.Sprintf("/boot/grub2/info%s", partition)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read firmware version from %s(%v)", path, err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("partition data is empty in %s", path)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no lines found in partition data")
	}

	for _, line := range lines {
		if !strings.HasPrefix(line, "firmware_version") {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid install date format: %s", line)
		}

		version := strings.TrimSpace(parts[1])
		return version, nil
	}

	return "", fmt.Errorf(
		"install date not found in partition data",
	)
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
