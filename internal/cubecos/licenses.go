package cubecos

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func ImportClusterLicense(licensePath string) error {
	dir := filepath.Dir(licensePath)
	base := filepath.Base(licensePath)
	_, err := exec.Command("hex_config", "sdk_run", "license_cluster_import", dir, base).Output()
	if err != nil {
		log.Errorf("failed to import licenses: %v", err)
		return err
	}
	return nil
}

func ImportNodeLicense(licensePath string) error {
	dir := filepath.Dir(licensePath)
	filename := strings.TrimSuffix(filepath.Base(licensePath), filepath.Ext(licensePath))

	_, err := exec.Command("hex_config", "sdk_run", "license_node_import", dir, filename).Output()
	if err != nil {
		log.Errorf("failed to import licenses: %v", err)
		return err
	}
	return nil
}

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
		licenses = append(licenses, genApiLicense(rawLicense))
	}

	return licenses
}

func genApiLicense(rawLicense definition.RawLicense) definition.License {
	issueDate, err := time.Parse("2006-01-02 15:04:05 MST", rawLicense.Date)
	if err != nil {
		rawLicense.Date = "unknown issue date"
	}

	expiry, err := time.Parse("2006-01-02 15:04:05 MST", rawLicense.Expiry)
	if err != nil {
		rawLicense.Expiry = "unknown expiry date"
	}

	return definition.License{
		Type:     rawLicense.Type,
		Hostname: rawLicense.Hostname,
		Serial:   rawLicense.Serial,
		Issue: definition.Issue{
			By:       rawLicense.IssueBy,
			To:       rawLicense.IssueTo,
			Hardware: rawLicense.Hardware,
			Date:     issueDate.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
		},
		Expiry: definition.Expiry{
			Date: expiry.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
			Days: rawLicense.Days,
		},
	}
}
