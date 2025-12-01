package cubecos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func ListFirmwares() ([]firmwares.Firmware, error) {
	update, err := parseUpdateHistory()
	if err != nil {
		return nil, err
	}

	firmwares := convertHistoryToFirmwares(update)
	firmwares = removeFreshInstalledFirmwares(firmwares)
	appendUninstalledFirmwares(&firmwares)

	return firmwares, nil
}

// note:
// please DO NOT use exec.CommandContext with timeout for hex_install
// because the duration of firmware upgrade is not predictable from CubeCOS, and it may take a long time to complete.
// use timeout might makes the situation to be worse.
func UpgradeFirmware(req *firmwares.ReqOpts) error {
	out, err := exec.Command("hex_install", "-v", "update", req.PkgPath).CombinedOutput()
	if err != nil {
		errDesc := strings.ReplaceAll(string(out), "\n", " ")
		log.Errorf("firmwares: failed to execute firmware upgrade %s(%s %s)", req.Version, err, errDesc)
		code, stderr := getUpdateFirmwareStatus()
		return fmt.Errorf("FRW0000%dE: %s", code, stderr)
	}

	if !IsHexSuccessful(err) {
		err := fmt.Errorf("%v %s", err, string(out))
		log.Errorf("firmwares: failed to upgrade firmware(%v)", err)
		code, stderr := getUpdateFirmwareStatus()
		return fmt.Errorf("FRW0000%dE: %s", code, stderr)
	}

	log.Infof("firmwares: %s", string(out))
	return nil
}

func getUpdateFirmwareErrCode(err error) int {
	code := GetCmdReturnCode(err)
	switch code {
	case 1:
		return 4
	case 2:
		return 5
	case 3:
		return 6
	default:
		return 0
	}
}

func getUpdateFirmwareStatus() (int, string) {
	out, err := exec.Command("hex_sdk", "-v", "stats_partition").CombinedOutput()
	if err != nil {
		intgErr := genIntegrationErr("firmware fetch status exec failure")
		log.Errorf("firmwares: %s (%s)", intgErr.Error(), string(out))
		return getUpdateFirmwareErrCode(err), string(out)
	}

	if !IsHexSuccessful(err) {
		intgErr := genIntegrationErr("firmware fetch status output failure")
		log.Errorf("firmwares: %s (%s)", intgErr.Error(), string(out))
		return getUpdateFirmwareErrCode(err), string(out)
	}

	return 0, string(out)
}

func GetUpdateInterruptedNode() (*nodes.Node, error) {
	return nil, fmt.Errorf("waiting COS to provide the SDK, so not implemented yet")
}

func GetBoostrappingProgress() ([]firmwares.BoostrappingStatus, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(180))
	defer cancel()

	out, err := exec.CommandContext(ctx, "hex_sdk", "-f", "json", "stats_bootstrap").CombinedOutput()
	if err != nil {
		err := genIntegrationErr("boostrapping progress exec failure")
		log.Errorf("firmwares: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("boostrapping progress output failure")
		log.Errorf("firmwares: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	var status []firmwares.BoostrappingStatus
	err = json.Unmarshal(out, &status)
	if err != nil {
		err := genIntegrationErr("boostrapping progress output parsing failure")
		log.Errorf("firmwares: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return status, nil
}

func parseUpdateHistory() (*firmwares.Upadte, error) {
	data, err := os.ReadFile(firmwares.UpdateHistory)
	if err != nil {
		log.Errorf("firmwares: failed to read update history file %s (%v)", firmwares.UpdateHistory, err)
		return nil, err
	}

	update := &firmwares.Upadte{}
	err = yaml.Unmarshal(data, update)
	if err != nil {
		log.Errorf("firmwares: failed to unmarshal update history file %s (%v)", firmwares.UpdateHistory, err)
		return nil, err
	}

	return update, nil
}

func convertHistoryToFirmwares(update *firmwares.Upadte) []firmwares.Firmware {
	firmwaresList := make([]firmwares.Firmware, 0, len(update.History))

	for _, raw := range update.History {
		date := convertRawTime(time.FormatFirmware, raw.CreatedAt)
		dayBaseDate := convertRawTimeToDayBaseDate(raw.BuiltAt)
		firmwaresList = append(firmwaresList, firmwares.Firmware{
			Version:      convertFirmwareVersion(raw.Version, dayBaseDate),
			ReleaseNotes: convertReleaseNotes(raw.Version, raw.Variant, date),
			UpdatedAt:    date,
			Status: status.Firmware{
				Current:     status.Succeeded,
				IsRemovable: false,
			},
		})
	}

	return firmwaresList
}

func convertRawTimeToDayBaseDate(rawTime string) string {
	segments := strings.Split(rawTime, " ")
	if len(segments) < 2 {
		return ""
	}

	return segments[0]
}

func removeFreshInstalledFirmwares(firmwares []firmwares.Firmware) []firmwares.Firmware {
	if len(firmwares) > 0 {
		firmwares = firmwares[1:]
	}

	return firmwares
}

func appendUninstalledFirmwares(list *[]firmwares.Firmware) {
	isInstallted := map[string]bool{}
	for _, firmware := range *list {
		isInstallted[firmware.Version] = true
	}

	entries, err := os.ReadDir(firmwares.UpdateDir)
	if err != nil {
		log.Errorf("firmwares: failed to read update directory %s (%v)", firmwares.UpdateDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".pkg") {
			continue
		}

		firmware, err := ConvertPkgNameToFirmware(entry.Name())
		if err != nil {
			continue
		}

		if !isInstallted[firmware.Version] {
			(*list) = append(*list, *firmware)
		}
	}
}

func convertRawTime(layout, rawTime string) string {
	t, err := ostime.ParseInLocation(layout, rawTime, time.LocalFixedZone)
	if err != nil {
		log.Errorf("firmwares: failed to parse time %s (%v)", rawTime, err)
		return ""
	}

	return time.RFC3339Z(t)
}

func ConvertPkgNameToFirmware(pkgname string) (*firmwares.Firmware, error) {
	pkgname = strings.TrimSuffix(pkgname, ".pkg")
	segment := strings.Split(pkgname, "_")
	if len(segment) < 3 {
		err := fmt.Errorf("invalid firmware package name: %s", pkgname)
		log.Errorf("firmwares: %v", err)
		return nil, err
	}

	date := convertRawTime(time.FormatFirmwarePkg, segment[2])
	return &firmwares.Firmware{
		Version:      convertFirmwareVersion(segment[1], segment[2]),
		ReleaseNotes: convertReleaseNotes(segment[1], segment[3], date),
		UpdatedAt:    date,
		Status: status.Firmware{
			Current:     status.Available,
			IsUpdatable: true,
			IsRemovable: true,
		},
	}, nil
}

func convertFirmwareVersion(version, date string) string {
	return fmt.Sprintf("Cube Appliance %s %s", version, date)
}

func convertReleaseNotes(version, variant, date string) string {
	return fmt.Sprintf("The CubeCOS %s(%s) firmware release since %s", version, variant, date)
}
