package cubecos

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
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

	raws, err := parseRawLicenses(b)
	if err != nil {
		log.Errorf("licenses: failed to parse raw licenses: %v", err)
		return nil, err
	}

	return parseLicenses(raws), nil
}

func parseRawLicenses(b []byte) ([]definition.RawLicense, error) {
	raws := []definition.RawLicense{}
	err := json.Unmarshal(b, &raws)
	if err != nil {
		return nil, err
	}
	if len(raws) <= 0 {
		return nil, nil
	}

	return raws, nil
}

func parseLicenses(raws []definition.RawLicense) []definition.License {
	licenses := convertToLicenses(raws)
	return aggregateLicenses(licenses)
}

func convertToLicenses(raws []definition.RawLicense) []definition.License {
	licenses := []definition.License{}
	for _, raw := range raws {
		licenses = append(
			licenses,
			parseLicense(raw),
		)
	}

	return licenses
}

func aggregateLicenses(licenses []definition.License) []definition.License {
	mergedLicenses := []definition.License{}
	for _, license := range genLicenseMap(licenses) {
		mergedLicenses = append(mergedLicenses, license)
	}

	return mergedLicenses
}

func genLicenseMap(licenses []definition.License) map[string]definition.License {
	licenseMap := map[string]definition.License{}
	for _, license := range licenses {
		key := license.Key()
		mappedLicense, found := licenseMap[key]
		if !found {
			licenseMap[key] = license
			continue
		}

		mappedLicense.Hosts = slices.Concat(mappedLicense.Hosts, license.Hosts)
		licenseMap[key] = mappedLicense
	}

	return licenseMap
}

func parseLicense(raw definition.RawLicense) definition.License {
	return definition.License{
		Type:    raw.Type,
		Hosts:   []string{raw.Hostname},
		Product: parseProduct(raw.Product),
		Serial:  raw.Serial,
		Issue:   parseIssue(raw),
		Expiry:  parseExpiry(raw),
		Status:  parseStatus(raw),
	}
}

func parseProduct(raw definition.Product) definition.Product {
	if raw.Name == "" {
		raw.Name = "CubeCOS"
	}

	return definition.Product{
		Name:     raw.Name,
		Features: parseFeatures(raw),
	}
}

func parseFeatures(raw definition.Product) []string {
	if len(raw.Features) > 0 {
		return raw.Features
	}

	switch raw.Name {
	case "CubeCOS":
		raw.Features = []string{"virtualization", "kubernetes"}
	case "CubeCMP":
		raw.Features = []string{"all"}
	}

	return raw.Features
}

func parseIssue(raw definition.RawLicense) definition.Issue {
	issue, err := time.Parse("2006-01-02 15:04:05 MST", raw.Date)
	if err != nil {
		raw.Date = "unknown issue date"
	}

	return definition.Issue{
		By:       raw.IssueBy,
		To:       raw.IssueTo,
		Hardware: raw.Hardware,
		Date:     issue.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
	}
}

func parseExpiry(raw definition.RawLicense) definition.Expiry {
	expiry, err := time.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err != nil {
		raw.Expiry = "unknown expiry date"
	}

	return definition.Expiry{
		Date: expiry.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
		Days: raw.Days,
	}
}

// note:
// the reason to assign time.Now().Local() to expiry if expiry is invalid is that
// the expiry field shouldn't be invalid in whatever case, if it's invalid, then there must be something wrong
// during signing process, we should raise the unexpected symptom and block the further process.
func parseStatus(raw definition.RawLicense) status.License {
	expiry, err := time.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err != nil {
		expiry = time.Now().Local()
	}

	if time.Now().After(expiry) {
		return status.License{
			Current: "expired",
		}
	}

	if time.Now().AddDate(0, 0, 30).After(expiry) {
		return status.License{
			Current:    "expiring",
			IsExpiring: true,
		}
	}

	return status.License{
		Current:    "ok",
		IsExpiring: false,
	}
}
