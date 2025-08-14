package cubecos

import (
	"os"
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

	return convertToFirmwares(update), nil
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
			Version:      raw.Version,
			ReleaseNotes: raw.Image,
			UpdatedAt:    convertRawTime(raw.CreatedAt),
			Status:       status.Firmware{Current: status.Updated},
		})
	}

	return firmwaresList
}

func convertRawTime(rawTime string) string {
	t, err := ostime.Parse(time.FormatFirmware, rawTime)
	if err != nil {
		log.Errorf("firmwares: failed to parse time %s (%v)", rawTime, err)
		return ""
	}

	return time.LocalRFC3339(t)
}
