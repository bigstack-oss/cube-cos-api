package cubecos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/zip"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	json "github.com/json-iterator/go"
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

func IsLicenseFile(file string) bool {
	return strings.Contains(file, "license.")
}

func SyncSourceLicense() {
	b, err := exec.Command("hex_config", "sdk_run", "-f", "json", "license_cluster_show").Output()
	if err != nil {
		log.Errorf("licenses: licenses: failed to list licenses: %v", err)
		return
	}

	raws, err := parseRawLicenses(b)
	if err != nil {
		log.Errorf("licenses: licenses: failed to parse raw licenses: %v", err)
		return
	}

	list := parseLicenses(raws)
	licenses.SetList(list)
}

func VerifyLicense(file string) (*licenses.Verification, error) {
	defer os.Remove(file)
	checkInfo := checkImportLicense(file)
	dat, err := parseLicenseDat(file, checkInfo)
	if err != nil {
		log.Errorf("licenses: failed to parse license: %v", err)
		return nil, err
	}

	return &licenses.Verification{
		License:     *dat,
		EffectNodes: getLicenseEffectNodes(dat.Issue.Hardware),
	}, nil
}

// note:
// currently, the COS license import result is not clear by identifying the return code
// because it will still return 0 even the result is not ok.
// see ticket to know more https://github.com/bigstack-oss/cubecos/issues/29
func ImportClusterLicense(licensePath string) error {
	dir, base := getDirAndLicenseName(licensePath)
	out, err := exec.Command("hex_config", "sdk_run", "license_cluster_import", dir, base).Output()
	if err != nil {
		log.Errorf("licenses: failed to import licenses: %v (%s)", err, string(out))
		return err
	}

	return nil
}

func ImportNodeLicense(licensePath string) error {
	dir := filepath.Dir(licensePath)
	filename := strings.TrimSuffix(filepath.Base(licensePath), filepath.Ext(licensePath))
	out, err := exec.Command("hex_config", "sdk_run", "license_node_import", dir, filename).Output()
	if err != nil {
		log.Errorf("licenses: failed to import licenses: %v(%s)", err, string(out))
		return err
	}

	return nil
}

func ListLicenses() ([]licenses.License, error) {
	licenses := licenses.List()
	if len(licenses) == 0 {
		return nil, errors.ErrLicensesNotFound
	}

	return licenses, nil
}

func parseRawLicenses(b []byte) ([]licenses.Raw, error) {
	raws := []licenses.Raw{}
	err := json.Unmarshal(b, &raws)
	if err != nil {
		return nil, err
	}
	if len(raws) <= 0 {
		return nil, nil
	}

	return raws, nil
}

func parseLicenses(raws []licenses.Raw) []licenses.License {
	licenses := convertToLicenses(raws)
	return aggregateLicenses(licenses)
}

func convertToLicenses(raws []licenses.Raw) []licenses.License {
	licenses := []licenses.License{}
	for _, raw := range raws {
		if raw.IsUnlicense() {
			continue
		}

		licenses = append(
			licenses,
			parseLicense(raw),
		)
	}

	return licenses
}

func aggregateLicenses(list []licenses.License) []licenses.License {
	mergedLicenses := []licenses.License{}
	for _, license := range genLicenseMap(list) {
		mergedLicenses = append(mergedLicenses, license)
	}

	return mergedLicenses
}

func genLicenseMap(list []licenses.License) map[string]licenses.License {
	licenseMap := map[string]licenses.License{}
	for _, license := range list {
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

func parseLicense(raw licenses.Raw) licenses.License {
	return licenses.License{
		Name:        parseName(raw),
		Type:        raw.Type,
		Hosts:       []string{raw.Hostname},
		Product:     parseProduct(raw),
		Serial:      raw.Serial,
		Quantity:    parseQuantity(raw),
		SupportPlan: parseSupportPlan(raw),
		Issue:       parseIssue(raw),
		Expiry:      parseExpiry(raw),
		Status:      parseStatus(raw),
	}
}

func parseName(raw licenses.Raw) string {
	name := "CubeOS License"
	if raw.Name != "" {
		name = raw.Name
	}

	return name
}

func parseProduct(raw licenses.Raw) licenses.Product {
	name := licenses.CubeCOS
	if raw.Product != "" {
		name = raw.Product
	}

	feature := licenses.NA
	if raw.Feature != "" {
		feature = raw.Feature
	}

	return licenses.Product{
		Name:    name,
		Feature: feature,
	}
}

func parseQuantity(raw licenses.Raw) string {
	quantity := licenses.NA
	if raw.Quantity != "" {
		quantity = raw.Quantity
	}

	return quantity
}

func parseSupportPlan(raw licenses.Raw) string {
	supportPlan := licenses.NA
	if raw.SLA != "" {
		supportPlan = raw.SLA
	}

	return supportPlan
}

func parseIssue(raw licenses.Raw) licenses.Issue {
	date := ""
	issue, err := ostime.Parse("2006-01-02 15:04:05 MST", raw.Date)
	if err == nil {
		date = issue.In(time.LocalFixedZone).Format(time.FormatRFC3339)
	}

	return licenses.Issue{
		By:       raw.IssueBy,
		To:       raw.IssueTo,
		Hardware: raw.Hardware,
		Date:     date,
	}
}

func parseExpiry(raw licenses.Raw) licenses.Expiry {
	date := ""
	expiry, err := ostime.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err == nil {
		date = expiry.In(time.LocalFixedZone).Format(time.FormatRFC3339)
	}

	days := raw.Days
	s := parseStatus(raw)
	if s.Current == status.Expired {
		days = int(expiry.Sub(ostime.Now().Local()).Hours() / 24)
	}

	return licenses.Expiry{
		Date: date,
		Days: days,
	}
}

// note:
// the reason to assign time.Now().Local() to expiry if expiry is that
// the expiry field shouldn't be invalid in whatever case, if it's invalid, then there must be something wrong
// during signing process, we should raise the unexpected symptom and block the further process.
func parseStatus(raw licenses.Raw) status.License {
	if raw.Expiry == "" {
		return status.License{
			Current: status.Unlicense,
		}
	}

	expiry, err := ostime.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err != nil {
		expiry = ostime.Now().Local()
	}

	if ostime.Now().After(expiry) {
		return status.License{
			Current: status.Expired,
		}
	}

	if ostime.Now().AddDate(0, 0, 30).After(expiry) {
		return status.License{
			Current:    status.Valid,
			IsExpiring: true,
		}
	}

	return status.License{
		Current:    status.Valid,
		IsExpiring: false,
	}
}

func checkLicenseErr(err error) error {
	if err == nil {
		return nil
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return errors.ErrLicenseInternalSystemFailure
	}

	if result.ExitCode() == LicenseValid {
		return nil
	}

	switch result.ExitCode() {
	case LicenseExpired:
		return errors.ErrLicenseAlreadyExpired
	case LicenseNotInstalled:
		return errors.ErrLicenseNotInstalled
	case LicenseInvalidHardware:
		return errors.ErrLicenseInvalidHardware
	case LicenseInvalidSignature:
		return errors.ErrLicenseInvalidSignature
	case LicenseSysytemCompromised:
		return errors.ErrLicenseSystemCompromised
	}

	return errors.ErrLicenseUnknownStatus
}

func checkImportLicense(license string) error {
	dir, file := getDirAndLicenseName(license)
	err := zip.DecompressFromTo(license, dir)
	if err != nil {
		log.Errorf("licenses: failed to decompress license: %v", err)
		return err
	}

	_, err = exec.Command("hex_config", "license_check", "def", filepath.Join(dir, file)).Output()
	return checkLicenseErr(err)
}

func parseLicenseDat(file string, checkInfo error) (*licenses.License, error) {
	dir, name := getDirAndLicenseName(file)
	datFile, err := os.Open(filepath.Join(dir, fmt.Sprintf("%s.dat", name)))
	if err != nil {
		return nil, err
	}

	defer datFile.Close()
	licenseDat := &licenses.License{}
	setLicenseDat(datFile, licenseDat)
	setLicenseDatStatus(licenseDat, checkInfo)
	return licenseDat, nil
}

func getDirAndLicenseName(license string) (string, string) {
	return filepath.Dir(license), strings.TrimSuffix(filepath.Base(license), filepath.Ext(license))
}

func setLicenseDat(datFile *os.File, licenseDat *licenses.License) {
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
		log.Errorf("licenses: failed to read license dat file: %v", err)
		return
	}
}

func setLicenseDatStatus(licenseDat *licenses.License, checkInfo error) {
	if checkInfo == nil {
		licenseDat.SetValid()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseNotInstalled) {
		licenseDat.SetValid()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseAlreadyExpired) {
		licenseDat.SetExpired()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseInvalidHardware) {
		licenseDat.InitInvalidHardware()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseInvalidSignature) {
		licenseDat.InitInvalidSignature()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseSystemCompromised) {
		licenseDat.SetCompromised()
		return
	}
}

func getLicenseEffectNodes(hardwareInfo string) []licenses.Node {
	list := nodes.List()
	if isLicenseForAllNodes(hardwareInfo) {
		return convertToLicenseNodes(list)
	}

	hardwareSerials := strings.Split(hardwareInfo, ",")
	effectNodes := []nodes.Node{}
	for _, node := range list {
		if node.MatchHardwareSerial(hardwareSerials) {
			effectNodes = append(effectNodes, node)
		}
	}

	return convertToLicenseNodes(effectNodes)
}

func isLicenseForAllNodes(hardwareInfo string) bool {
	return strings.Contains(hardwareInfo, "*")
}

func convertToLicenseNodes(list []nodes.Node) []licenses.Node {
	nodes := []licenses.Node{}
	for _, node := range list {
		license := getNodeLicense(node.Hostname)
		if !license.IsValid() {
			license.Status.Current = status.Unlicense
		}

		nodes = append(
			nodes,
			licenses.Node{
				Name:   node.Hostname,
				Role:   node.Role,
				Expiry: license.Expiry,
				Status: license.Status,
			},
		)
	}

	return nodes
}

func getNodeLicense(nodeName string) licenses.License {
	list, err := ListLicenses()
	if err != nil {
		log.Errorf("licenses: failed to get license by node name: %v", err)
		return licenses.License{}
	}

	for _, license := range list {
		if slices.Contains(license.Hosts, nodeName) {
			return license
		}
	}

	return licenses.License{}
}

func setValueToLicenseDat(licenseDat *licenses.License, parts []string) {
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
		licenseDat.Product.Feature = value
	case "quantity":
		licenseDat.Quantity = value
	case "sla":
		licenseDat.SupportPlan = value
	case "issue.date":
		licenseDat.Issue.Date = parseDatIssueDate(value)
	case "expiry.date":
		expiry, status := parseExpiryAndStatus(value)
		licenseDat.Expiry = expiry
		licenseDat.Status = status
	}
}

func parseDatIssueDate(value string) string {
	issue, err := ostime.Parse("2006-01-02 15:04:05 MST", value)
	if err != nil {
		return "unknown issue date"
	}

	return issue.In(time.LocalFixedZone).Format(time.FormatRFC3339)
}

func parseExpiryAndStatus(value string) (licenses.Expiry, status.License) {
	expiry, err := ostime.Parse("2006-01-02 15:04:05 MST", value)
	if err != nil {
		return licenses.Expiry{
				Date: "unknown expiry date",
				Days: 0,
			}, status.License{
				Current: status.Expired,
			}
	}

	licenseExpiry := licenses.Expiry{
		Date: expiry.In(time.LocalFixedZone).Format(time.FormatRFC3339),
		Days: int(expiry.Sub(ostime.Now().Local()).Hours() / 24),
	}

	if ostime.Now().After(expiry) {
		return licenseExpiry, status.License{Current: status.Expired}
	}

	if ostime.Now().AddDate(0, 0, 30).After(expiry) {
		return licenseExpiry, status.License{Current: status.Expairing, IsExpiring: true}
	}

	return licenseExpiry, status.License{Current: status.Ok, IsExpiring: false}
}
