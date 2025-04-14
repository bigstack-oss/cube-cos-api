package cubecos

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/zip"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

const (
	LicenseValid              = 0
	LicenseSysytemCompromised = -1
	LicenseInvalidSignature   = -2
	LicenseInvalidHardware    = -3
	LicenseNotInstalled       = -4
	LicenseExpired            = -5
)

func VerifyLicense(license string) (*definition.VerificationDetails, error) {
	defer os.Remove(license)
	err := checkImportLicense(license)
	if err != nil {
		log.Errorf("license: failed to import license: %v", err)
		return nil, err
	}

	dat, err := parseLicenseDat(license)
	if err != nil {
		log.Errorf("license: failed to parse license: %v", err)
		return nil, err
	}

	return &definition.VerificationDetails{
		License:     *dat,
		EffectNodes: getLicenseEffectNodes(dat.Issue.Hardware),
	}, nil
}

// note:
// currently, the COS license import result is not clear by identifying the return code
// because it will still return 0 even if the result is not ok.
// see ticket to know more https://github.com/bigstack-oss/cubecos/issues/29
func ImportClusterLicense(licensePath string) error {
	dir, base := getDirAndLicenseName(licensePath)
	out, err := exec.Command("hex_config", "sdk_run", "license_cluster_import", dir, base).Output()
	if err != nil {
		log.Errorf("license: failed to import licenses: %v (%s)", err, string(out))
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

	if result.ExitCode() >= LicenseValid {
		return nil
	}

	switch result.ExitCode() {
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

func checkImportLicense(license string) error {
	dir, file := getDirAndLicenseName(license)
	_, err := exec.Command("hex_config", "license_check", "def", filepath.Join(dir, file)).Output()
	return checkLicenseErr(err)
}

func parseLicenseDat(license string) (*definition.License, error) {
	dir, file := getDirAndLicenseName(license)
	err := zip.DecompressFromTo(license, dir)
	if err != nil {
		log.Errorf("license: failed to decompress license: %v", err)
		return nil, err
	}

	datFile, err := os.Open(filepath.Join(dir, fmt.Sprintf("%s.dat", file)))
	if err != nil {
		return nil, err
	}

	defer datFile.Close()
	licenseDat := &definition.License{}
	setLicenseDat(datFile, licenseDat)
	return licenseDat, nil
}

func getDirAndLicenseName(license string) (string, string) {
	return filepath.Dir(license), strings.TrimSuffix(filepath.Base(license), filepath.Ext(license))
}

func setLicenseDat(datFile *os.File, licenseDat *definition.License) {
	scanner := bufio.NewScanner(datFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isCommentOrBlank(line) {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if !isKeyValuePattern(parts) {
			continue
		}

		setValueToLicenseDat(licenseDat, parts)
	}

	err := scanner.Err()
	if err != nil {
		log.Errorf("license: failed to read license dat file: %v", err)
		return
	}
}

func getLicenseEffectNodes(hardwareInfo string) []definition.LicenseNode {
	nodes := definition.ListNodes()
	if isLicenseForAllNodes(hardwareInfo) {
		return convertToLicenseNodes(nodes)
	}

	hardwareSerials := strings.Split(hardwareInfo, ",")
	effectNodes := []definition.Node{}
	for _, node := range nodes {
		if node.MatchHardwareSerial(hardwareSerials) {
			effectNodes = append(effectNodes, node)
		}
	}

	return convertToLicenseNodes(effectNodes)
}

func isLicenseForAllNodes(hardwareInfo string) bool {
	return strings.Contains(hardwareInfo, "*")
}

func convertToLicenseNodes(nodes []definition.Node) []definition.LicenseNode {
	licenseNodes := []definition.LicenseNode{}
	for _, node := range nodes {
		license := getLicenseByNodeName(node.Hostname)
		licenseNodes = append(
			licenseNodes,
			definition.LicenseNode{
				Name:   node.Hostname,
				Role:   node.Role,
				Expiry: license.Expiry,
				Status: license.Status,
			},
		)
	}

	return licenseNodes
}

func getLicenseByNodeName(nodeName string) definition.License {
	licenses, err := ListLicenses()
	if err != nil {
		log.Errorf("license: failed to get license by node name: %v", err)
		return definition.License{}
	}

	for _, license := range licenses {
		if slices.Contains(license.Hosts, nodeName) {
			return license
		}
	}

	return definition.License{}
}

func setValueToLicenseDat(licenseDat *definition.License, parts []string) {
	key := parts[0]
	value := strings.TrimSpace(parts[1])

	switch key {
	case "license.name":
		licenseDat.Name = value
	case "license.type":
		licenseDat.Type = value
	case "issue.by":
		licenseDat.Issue.By = value
	case "issue.to":
		licenseDat.Issue.To = value
	case "issue.hardware":
		licenseDat.Issue.Hardware = value
	case "product":
		licenseDat.Product.Name = value
	case "feature":
		licenseDat.Product.Features = append(licenseDat.Product.Features, value)
	case "quantity":
		licenseDat.Quantity = parseLicenseDatQuantity(value)
	case "sla":
		licenseDat.ServiceLevelAgreement = value
	case "issue.date":
		licenseDat.Issue.Date = parseLicenseDatIssueDate(value)
	case "expiry.date":
		expiry, status := parseLicenseExpiryAndStatus(value)
		licenseDat.Expiry = expiry
		licenseDat.Status = status
	}
}

func parseLicenseDatQuantity(value string) definition.Quantity {
	quantity := definition.Quantity{Value: 0}
	val, err := strconv.Atoi(value)
	if err != nil {
		return quantity
	}

	quantity.Value = val
	return quantity
}

func parseLicenseDatIssueDate(value string) string {
	issue, err := time.Parse("2006-01-02 15:04:05 MST", value)
	if err != nil {
		return "unknown issue date"
	}

	return issue.In(definition.LocalTimeFixedZone).Format(time.RFC3339)
}

func parseLicenseExpiryAndStatus(value string) (definition.Expiry, status.License) {
	expiry, err := time.Parse("2006-01-02 15:04:05 MST", value)
	if err != nil {
		return definition.Expiry{
				Date: "unknown expiry date",
				Days: 0,
			}, status.License{
				Current: "expired",
			}
	}

	licenseExpiry := definition.Expiry{
		Date: expiry.In(definition.LocalTimeFixedZone).Format(time.RFC3339),
		Days: int(expiry.Sub(time.Now().Local()).Hours() / 24),
	}

	if time.Now().After(expiry) {
		return licenseExpiry, status.License{Current: "expired"}
	}

	if time.Now().AddDate(0, 0, 30).After(expiry) {
		return licenseExpiry, status.License{Current: "expiring", IsExpiring: true}
	}

	return licenseExpiry, status.License{Current: "ok", IsExpiring: false}
}
