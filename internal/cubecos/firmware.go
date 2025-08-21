package cubecos

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
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
		err := fmt.Errorf("failed to execute firmware upgrade %s(%v %s)", req.Version, err, string(out))
		log.Errorf("firmwares: %v", err)
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf("failed to upgrade firmware %s(%v %s)", req.Version, err, string(out))
	}

	return nil
}

func GetUpdateInterruptedNode() (*nodes.Node, error) {
	return nil, fmt.Errorf("waiting COS to provide the SDK, so not implemented yet")
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
		firmwaresList = append(firmwaresList, firmwares.Firmware{
			Version:      convertFirmwareVersion(raw.Version),
			ReleaseNotes: convertReleaseNotes(raw.Version, raw.Variant, date),
			UpdatedAt:    date,
			Status: status.Firmware{
				Current:     status.Updated,
				IsRemovable: false,
			},
		})
	}

	return firmwaresList
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

		firmware, err := convertPkgNameToFirmware(entry.Name())
		if err != nil {
			continue
		}

		if !isInstallted[firmware.Version] {
			(*list) = append(*list, *firmware)
		}
	}
}

func convertRawTime(layout, rawTime string) string {
	t, err := ostime.Parse(layout, rawTime)
	if err != nil {
		log.Errorf("firmwares: failed to parse time %s (%v)", rawTime, err)
		return ""
	}

	return time.LocalRFC3339(t)
}

func convertPkgNameToFirmware(pkgname string) (*firmwares.Firmware, error) {
	pkgname = strings.TrimSuffix(pkgname, ".pkg")
	segment := strings.Split(pkgname, "_")
	if len(segment) < 3 {
		err := fmt.Errorf("invalid firmware package name: %s", pkgname)
		log.Errorf("firmwares: %v", err)
		return nil, err
	}

	date := convertRawTime(time.FormatFirmwarePkg, segment[2])
	return &firmwares.Firmware{
		Version:      convertFirmwareVersion(segment[1]),
		ReleaseNotes: convertReleaseNotes(segment[1], segment[3], date),
		UpdatedAt:    date,
		Status: status.Firmware{
			Current:     status.Available,
			IsUpdatable: true,
			IsRemovable: true,
		},
	}, nil
}

func convertFirmwareVersion(version string) string {
	return fmt.Sprintf("Cube Appliance %s", version)
}

func convertReleaseNotes(version, variant, date string) string {
	return fmt.Sprintf("The CubeCOS %s(%s) firmware release since %s", version, variant, date)
}
