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

	timeFormatLicense = "2006-01-02 15:04:05 MST"
)

func IsLicenseFile(file string) bool {
	return strings.Contains(file, "license")
}

func SyncSourceLicense() {
	b, err := exec.Command("hex_sdk", "-f", "json", "license_cluster_show").Output()
	if err != nil {
		log.Errorf("licenses: licenses: failed to list licenses(%v)", err)
		return
	}

	raws, err := parseRawLicenses(b)
	if err != nil {
		log.Errorf("licenses: licenses: failed to parse raw licenses(%v)", err)
		return
	}

	list := parseLicenses(raws)
	licenses.SetList(list)
}

func VerifyLicense(file string) (*licenses.Verification, error) {
	defer os.Remove(file)
	dat, err := parseLicenseDat(file)
	if err != nil {
		log.Errorf("licenses: failed to parse license(%v)", err)
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
	dir, base := getLicenseDirAndName(licensePath)
	out, err := exec.Command("hex_config", "sdk_run", "license_cluster_import", dir, base).Output()
	if err == nil {
		return nil
	}

	log.Errorf("licenses: failed to import cluster license: %v(%s)", err, string(out))
	return checkLicenseErr(err)
}

func ImportNodeLicense(licensePath string) error {
	dir := filepath.Dir(licensePath)
	filename := strings.TrimSuffix(filepath.Base(licensePath), filepath.Ext(licensePath))
	out, err := exec.Command("hex_config", "sdk_run", "license_node_import", dir, filename).Output()
	if err == nil {
		return nil
	}

	log.Errorf("licenses: failed to import node license: %v(%s)", err, string(out))
	return checkLicenseErr(err)
}

func ListLicenses() []licenses.License {
	list := licenses.List()
	if len(list) == 0 {
		return []licenses.License{}
	}

	return list
}

func GetHostLicense(hostname string) licenses.License {
	list := ListLicenses()
	if licenses.IsNotInstalled(list) {
		return licenses.License{
			Status: status.License{
				Current: status.Unlicense,
			},
		}
	}

	for _, license := range list {
		if slices.Contains(license.Hosts, hostname) {
			license.Hosts = nil
			return license
		}
	}

	return licenses.License{}
}

func parseRawLicenses(b []byte) ([]licenses.Raw, error) {
	raws := []licenses.Raw{}
	err := json.Unmarshal(b, &raws)
	if err != nil {
		log.Errorf("licenses: failed to parse raw licenses(%v)", err)
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
		Name:        parseName(raw.Name),
		Type:        raw.Type,
		Hosts:       []string{raw.Hostname},
		Product:     parseProduct(raw),
		Serial:      raw.Serial,
		Quantity:    parseQuantity(raw),
		SupportPlan: parseSupportPlan(raw.SupportPlan),
		Issue:       parseIssue(raw),
		Expiry:      parseExpiry(raw),
		Status:      parseStatus(raw),
	}
}

func parseName(value string) string {
	name := "CubeOS License"
	if value != "" {
		name = value
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

func parseSupportPlan(value string) string {
	supportPlan := licenses.NA
	if value != "" {
		supportPlan = value
	}

	return supportPlan
}

func parseIssue(raw licenses.Raw) licenses.Issue {
	date := ""
	issue, err := ostime.Parse(timeFormatLicense, raw.Date)
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
	expiry, err := ostime.Parse(timeFormatLicense, raw.Expiry)
	if err == nil {
		date = expiry.In(time.LocalFixedZone).Format(time.FormatRFC3339)
	}

	return licenses.Expiry{
		Date: date,
		Days: parseExpiryDuration(expiry),
	}
}

func parseExpiryDuration(expiry ostime.Time) int {
	secondsAgo := expiry.Sub(ostime.Now().Local()).Seconds() / 86400.0
	if secondsAgo < 1 {
		return -1
	}

	return int(secondsAgo)
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

	expiry, err := ostime.Parse(timeFormatLicense, raw.Expiry)
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

func parseLicenseDat(file string) (*licenses.License, error) {
	err := unzipLicense(file)
	if err != nil {
		return nil, err
	}

	dir, name := getLicenseDirAndName(file)
	dat, err := os.Open(filepath.Join(dir, fmt.Sprintf("%s.dat", name)))
	if err != nil {
		return nil, err
	}

	defer dat.Close()
	license := &licenses.License{}
	setLicenseDat(dat, license)
	setLicenseDatStatus(
		license,
		checkImportLicense(file, *license),
	)

	return license, nil
}

func unzipLicense(license string) error {
	dir, _ := getLicenseDirAndName(license)
	err := zip.DecompressFromTo(license, dir)
	if err != nil {
		log.Errorf("licenses: failed to unzip license(%v)", err)
		return err
	}

	return nil
}

func checkImportLicense(file string, license licenses.License) error {
	dir, file := getLicenseDirAndName(file)
	product := "def"
	if license.Product.Name == licenses.CubeCMP {
		product = "cmp"
	}

	_, err := exec.Command("hex_config", "license_check", product, filepath.Join(dir, file)).Output()
	return checkLicenseErr(err)
}

func getLicenseDirAndName(license string) (string, string) {
	return filepath.Dir(license), strings.TrimSuffix(filepath.Base(license), filepath.Ext(license))
}

func setLicenseDat(datFile *os.File, license *licenses.License) {
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

		setValueToLicenseDat(license, parts)
	}

	err := scanner.Err()
	if err != nil {
		log.Errorf("licenses: failed to read license dat file(%v)", err)
		return
	}

	backfillValueForOldLicense(license)
}

func backfillValueForOldLicense(license *licenses.License) {
	if license.Name == "" {
		license.Name = licenses.CubeCOS
	}

	if license.Product.Name == "" {
		license.Product.Name = licenses.CubeCOS
	}

	if license.Product.Feature == "" {
		license.Product.Feature = licenses.NA
	}

	if license.SupportPlan == "" {
		license.SupportPlan = licenses.NA
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
	for _, license := range ListLicenses() {
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
		licenseDat.Name = parseName(value)
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
		licenseDat.Product.Feature = parseFeature(value)
	case "quantity":
		licenseDat.Quantity = value
	case "support.plan":
		licenseDat.SupportPlan = parseSupportPlan(value)
	case "issue.date":
		licenseDat.Issue.Date = parseDatIssueDate(value)
	case "expiry.date":
		expiry, status := parseExpiryAndStatus(value)
		licenseDat.Expiry = expiry
		licenseDat.Status = status
	}
}

func parseFeature(value string) string {
	feature := licenses.NA
	if value != "" {
		feature = value
	}

	return feature
}

func parseDatIssueDate(value string) string {
	issue, err := ostime.Parse(timeFormatLicense, value)
	if err != nil {
		return "unknown issue date"
	}

	return issue.In(time.LocalFixedZone).Format(time.FormatRFC3339)
}

func parseExpiryAndStatus(value string) (licenses.Expiry, status.License) {
	expiry, err := ostime.Parse(timeFormatLicense, value)
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
