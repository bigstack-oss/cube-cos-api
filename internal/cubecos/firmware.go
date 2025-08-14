package cubecos

import (
	"fmt"
	"os"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
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

	firmwares := convertToFirmwares(update)
	firmwares = removeFreshInstalledFirmwares(firmwares)
	appendUninstalledFirmwares(&firmwares)

	return firmwares, nil
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

func convertToFirmwares(update *firmwares.Upadte) []firmwares.Firmware {
	firmwaresList := make([]firmwares.Firmware, 0, len(update.History))

	for _, raw := range update.History {
		firmwaresList = append(firmwaresList, firmwares.Firmware{
			Version:   fmt.Sprintf("%s Appliance %s %s", raw.Image, raw.Version, raw.Variant),
			UpdatedAt: convertRawTime(time.FormatFirmware, raw.CreatedAt),
			Status:    status.Firmware{Current: status.Updated},
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

	return &firmwares.Firmware{
		Version:   fmt.Sprintf("%s Appliance %s %s", segment[0], segment[1], segment[3]),
		UpdatedAt: convertRawTime(time.FormatFirmwarePkg, segment[2]),
		Status: status.Firmware{
			Current:     status.Available,
			IsUpdatable: true,
		},
	}, nil
}
