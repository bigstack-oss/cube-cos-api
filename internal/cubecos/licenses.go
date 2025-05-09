package cubecos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/zip"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
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

	licenses := parseLicenses(raws)
	license.SetList(licenses)
}

func VerifyLicense(file string) (*license.Verification, error) {
	defer os.Remove(file)
	checkInfo := checkImportLicense(file)
	dat, err := parseLicenseDat(file, checkInfo)
	if err != nil {
		log.Errorf("licenses: failed to parse license: %v", err)
		return nil, err
	}

	return &license.Verification{
		Options:     *dat,
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

func ListLicenses() ([]license.Options, error) {
	licenses := license.List()
	if len(licenses) == 0 {
		return nil, errors.ErrLicensesNotFound
	}

	return licenses, nil
}

func parseRawLicenses(b []byte) ([]license.Raw, error) {
	raws := []license.Raw{}
	err := json.Unmarshal(b, &raws)
	if err != nil {
		return nil, err
	}
	if len(raws) <= 0 {
		return nil, nil
	}

	return raws, nil
}

func parseLicenses(raws []license.Raw) []license.Options {
	licenses := convertToLicenses(raws)
	return aggregateLicenses(licenses)
}

func convertToLicenses(raws []license.Raw) []license.Options {
	licenses := []license.Options{}
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

func aggregateLicenses(licenses []license.Options) []license.Options {
	mergedLicenses := []license.Options{}
	for _, license := range genLicenseMap(licenses) {
		mergedLicenses = append(mergedLicenses, license)
	}

	return mergedLicenses
}

func genLicenseMap(licenses []license.Options) map[string]license.Options {
	licenseMap := map[string]license.Options{}
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

func parseLicense(raw license.Raw) license.Options {
	return license.Options{
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

func parseName(raw license.Raw) string {
	name := "CubeOS License"
	if raw.Name != "" {
		name = raw.Name
	}

	return name
}

func parseProduct(raw license.Raw) license.Product {
	name := "CubeCOS"
	if raw.Product != "" {
		name = raw.Product
	}

	feature := "N/A"
	if raw.Feature != "" {
		feature = raw.Feature
	}

	return license.Product{
		Name:    name,
		Feature: feature,
	}
}

func parseQuantity(raw license.Raw) string {
	quantity := "N/A"
	if raw.Quantity != "" {
		quantity = raw.Quantity
	}

	return quantity
}

func parseSupportPlan(raw license.Raw) string {
	supportPlan := "N/A"
	if raw.SLA != "" {
		supportPlan = raw.SLA
	}

	return supportPlan
}

func parseIssue(raw license.Raw) license.Issue {
	date := ""
	issue, err := time.Parse("2006-01-02 15:04:05 MST", raw.Date)
	if err == nil {
		date = issue.In(v1.LocalTimeFixedZone).Format(time.RFC3339)
	}

	return license.Issue{
		By:       raw.IssueBy,
		To:       raw.IssueTo,
		Hardware: raw.Hardware,
		Date:     date,
	}
}

func parseExpiry(raw license.Raw) license.Expiry {
	date := ""
	expiry, err := time.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err == nil {
		date = expiry.In(v1.LocalTimeFixedZone).Format(time.RFC3339)
	}

	return license.Expiry{
		Date: date,
		Days: raw.Days,
	}
}

// note:
// the reason to assign time.Now().Local() to expiry if expiry is that
// the expiry field shouldn't be invalid in whatever case, if it's invalid, then there must be something wrong
// during signing process, we should raise the unexpected symptom and block the further process.
func parseStatus(raw license.Raw) status.License {
	if raw.Expiry == "" {
		return status.License{
			Current: status.Unlicense,
		}
	}

	expiry, err := time.Parse("2006-01-02 15:04:05 MST", raw.Expiry)
	if err != nil {
		expiry = time.Now().Local()
	}

	if time.Now().After(expiry) {
		return status.License{
			Current: status.Expired,
		}
	}

	if time.Now().AddDate(0, 0, 30).After(expiry) {
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
		return errors.ErrLicenseSysytemCompromised
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

func parseLicenseDat(file string, checkInfo error) (*license.Options, error) {
	dir, name := getDirAndLicenseName(file)
	datFile, err := os.Open(filepath.Join(dir, fmt.Sprintf("%s.dat", name)))
	if err != nil {
		return nil, err
	}

	defer datFile.Close()
	licenseDat := &license.Options{}
	setLicenseDat(datFile, licenseDat)
	setLicenseDatStatus(licenseDat, checkInfo)
	return licenseDat, nil
}

func getDirAndLicenseName(license string) (string, string) {
	return filepath.Dir(license), strings.TrimSuffix(filepath.Base(license), filepath.Ext(license))
}

func setLicenseDat(datFile *os.File, licenseDat *license.Options) {
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

func setLicenseDatStatus(licenseDat *license.Options, checkInfo error) {
	if checkInfo == nil {
		licenseDat.InitValidStatus()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseNotInstalled) {
		licenseDat.InitValidStatus()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseAlreadyExpired) {
		licenseDat.InitExpiredStatus()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseInvalidHardware) {
		licenseDat.InitInvalidHardwareStatus()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseInvalidSignature) {
		licenseDat.InitInvalidSignatureStatus()
		return
	}

	if errors.Is(checkInfo, errors.ErrLicenseSysytemCompromised) {
		licenseDat.InitCompromisedStatus()
		return
	}
}

func getLicenseEffectNodes(hardwareInfo string) []license.Node {
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

func convertToLicenseNodes(nodes []nodes.Node) []license.Node {
	licenseNodes := []license.Node{}
	for _, node := range nodes {
		l := getLicenseByNodeName(node.Hostname)
		if !l.IsValid() {
			l.Status.Current = "unlicense"
		}

		licenseNodes = append(
			licenseNodes,
			license.Node{
				Name:   node.Hostname,
				Role:   node.Role,
				Expiry: l.Expiry,
				Status: l.Status,
			},
		)
	}

	return licenseNodes
}

func getLicenseByNodeName(nodeName string) license.Options {
	licenses, err := ListLicenses()
	if err != nil {
		log.Errorf("licenses: failed to get license by node name: %v", err)
		return license.Options{}
	}

	for _, license := range licenses {
		if slices.Contains(license.Hosts, nodeName) {
			return license
		}
	}

	return license.Options{}
}

func setValueToLicenseDat(licenseDat *license.Options, parts []string) {
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
		licenseDat.Issue.Date = parseLicenseDatIssueDate(value)
	case "expiry.date":
		expiry, status := ParseLicenseExpiryAndStatus(value)
		licenseDat.Expiry = expiry
		licenseDat.Status = status
	}
}

func parseLicenseDatIssueDate(value string) string {
	issue, err := time.Parse("2006-01-02 15:04:05 MST", value)
	if err != nil {
		return "unknown issue date"
	}

	return issue.In(v1.LocalTimeFixedZone).Format(time.RFC3339)
}

func ParseLicenseExpiryAndStatus(value string) (license.Expiry, status.License) {
	expiry, err := time.Parse("2006-01-02 15:04:05 MST", value)
	if err != nil {
		return license.Expiry{
				Date: "unknown expiry date",
				Days: 0,
			}, status.License{
				Current: "expired",
			}
	}

	licenseExpiry := license.Expiry{
		Date: expiry.In(v1.LocalTimeFixedZone).Format(time.RFC3339),
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
