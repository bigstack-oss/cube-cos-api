package cubecos

import (
	"encoding/json"
	"os/exec"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func ListLicenses() ([]definition.License, error) {
	b, err := exec.Command("hex_config", "sdk_run", "-f", "json", "license_cluster_show").Output()
	if err != nil {
		log.Errorf("failed to list licenses: %v", err)
		return nil, err
	}

	rawLicenses := []definition.RawLicense{}
	err = json.Unmarshal(b, &rawLicenses)
	if err != nil {
		log.Errorf("failed to unmarshal licenses: %v", err)
		return nil, err
	}
	if len(rawLicenses) <= 0 {
		return nil, nil
	}

	return convertRawLicensesToApiLicenses(rawLicenses), nil
}

func convertRawLicensesToApiLicenses(rawLicenses []definition.RawLicense) []definition.License {
	licenses := []definition.License{}

	for _, rawLicense := range rawLicenses {
		licenses = append(
			licenses,
			definition.License{
				Type:     rawLicense.Type,
				Hostname: rawLicense.Hostname,
				Serial:   rawLicense.Serial,
				Issue: definition.Issue{
					By:       rawLicense.IssueBy,
					To:       rawLicense.IssueTo,
					Hardware: rawLicense.Hardware,
					Date:     rawLicense.Date,
				},
				Expiry: definition.Expiry{
					Date: rawLicense.Expiry,
					Days: rawLicense.Days,
				},
			},
		)
	}

	return licenses
}
