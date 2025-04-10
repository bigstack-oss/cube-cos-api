package cubecos

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

const (
	LicenseValid              = 1
	LicenseExpired            = 251
	LicenseNotInstalled       = 252
	LicenseInvalidHardware    = 253
	LicenseInvalidSignature   = 254
	LicenseSysytemCompromised = 255
)

func VerifyLicense(license string) (*definition.License, error) {
	defer os.Remove(license)

	dir, file := getDirAndLicenseName(license)
	_, err := exec.Command("hex_sdk", "license_import_check", dir, file).Output()
	err = checkLicenseErr(err)
	if err != nil {
		log.Errorf("license: failed to verify license: %v", err)
		return nil, err
	}

	return nil, nil
}

func getDirAndLicenseName(license string) (string, string) {
	return filepath.Dir(license), strings.TrimSuffix(filepath.Base(license), filepath.Ext(license))
}

func ImportClusterLicense(licensePath string) error {
	dir := filepath.Dir(licensePath)
	base := filepath.Base(licensePath)
	_, err := exec.Command("hex_config", "sdk_run", "license_cluster_import", dir, base).Output()
	if err != nil {
		log.Errorf("license: failed to import licenses: %v", err)
		return err
	}

	return nil
}

func ImportNodeLicense(licensePath string) error {
	dir := filepath.Dir(licensePath)
	filename := strings.TrimSuffix(filepath.Base(licensePath), filepath.Ext(licensePath))

	_, err := exec.Command("hex_config", "sdk_run", "license_node_import", dir, filename).Output()
	if err != nil {
		log.Errorf("license: failed to import licenses: %v", err)
		return err
	}

	return nil
}

func ListLicenses() ([]definition.License, error) {
	b, err := exec.Command("hex_config", "sdk_run", "-f", "json", "license_cluster_show").Output()
	if err != nil {
		log.Errorf("license: licenses: failed to list licenses: %v", err)
		return nil, err
	}

	raws, err := parseRawLicenses(b)
	if err != nil {
		log.Errorf("license: licenses: failed to parse raw licenses: %v", err)
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
		Name:                  raw.Name,
		Type:                  raw.Type,
		Hosts:                 []string{raw.Hostname},
		Product:               parseProduct(raw),
		Serial:                raw.Serial,
		ServiceLevelAgreement: raw.SLA,
		Issue:                 parseIssue(raw),
		Expiry:                parseExpiry(raw),
		Status:                parseStatus(raw),
	}
}

func parseProduct(raw definition.RawLicense) definition.Product {
	if raw.Product == "" {
		raw.Product = "CubeCOS"
	}

	return definition.Product{
		Name:     raw.Product,
		Features: parseFeatures(raw),
	}
}

func parseFeatures(raw definition.RawLicense) []string {
	features := []string{}

	switch raw.Product {
	case "CubeCOS":
		features = []string{"virtualization", "kubernetes"}
	case "CubeCMP":
		features = []string{"all"}
	}

	return features
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

func checkLicenseErr(err error) error {
	if err == nil {
		return nil
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return errors.New("internal license system error")
	}

	switch result.ExitCode() {
	case LicenseValid:
		return nil
	case LicenseExpired:
		return errors.New("license is already expired")
	case LicenseNotInstalled:
		return errors.New("license is not installed")
	case LicenseInvalidHardware:
		return errors.New("license's hardware serial is not matched with the current system")
	case LicenseInvalidSignature:
		return errors.New("license's signature is invalid")
	case LicenseSysytemCompromised:
		return errors.New("license system is compromised")
	}

	return errors.New("unknown license status")
}
