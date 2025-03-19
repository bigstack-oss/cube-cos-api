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
		log.Errorf("licenses: failed to list licenses: %v", err)
		return nil, err
	}

	raws := []definition.RawLicense{}
	err = json.Unmarshal(b, &raws)
	if err != nil {
		log.Errorf("licenses: failed to unmarshal licenses: %v", err)
		return nil, err
	}
	if len(raws) <= 0 {
		return nil, nil
	}

	return parseLicenses(raws), nil
}

func parseLicenses(raws []definition.RawLicense) []definition.License {
	licenses := []definition.License{}
	for _, raw := range raws {
		licenses = append(
			licenses,
			parseLicense(raw),
		)
	}

	return licenses
}

func parseLicense(raw definition.RawLicense) definition.License {
	issue, err := time.Parse("2006-01-02 15:04:05 MST", raw.Date)
	if err != nil {
		raw.Date = "unknown issue date"
	}

	expiry, err := time.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err != nil {
		raw.Expiry = "unknown expiry date"
	}

	return definition.License{
		Type:     raw.Type,
		Hostname: raw.Hostname,
		Serial:   raw.Serial,
		Issue: definition.Issue{
			By:       raw.IssueBy,
			To:       raw.IssueTo,
			Hardware: raw.Hardware,
			Date:     issue.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
		},
		Expiry: definition.Expiry{
			Date: expiry.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
			Days: raw.Days,
		},
	}
}
