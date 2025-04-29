package v1

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
)

const (
	Licenses   = "licenses"
	LicenseDir = "/etc/update"
)

var (
	licenseSearcher bleve.Index
	license         = []License{}
	updateLicense   sync.Mutex
)

type VerificationDetails struct {
	License     `json:"license" yaml:"license" bson:"license"`
	EffectNodes []LicenseNode `json:"effectNodes" yaml:"effectNodes" bson:"effectNodes"`
}

type LicenseNode struct {
	Name   string `json:"name" yaml:"name" bson:"name"`
	Role   string `json:"role" yaml:"role" bson:"role"`
	Expiry `json:"expiry" yaml:"expiry" bson:"expiry"`
	Status status.License `json:"status" yaml:"status" bson:"status"`
}

type License struct {
	Name        string   `json:"name" yaml:"name" bson:"name"`
	Type        string   `json:"type" yaml:"type" bson:"type"`
	Hostname    string   `json:"hostname,omitzero" yaml:"hostname" bson:"hostname"`
	Hosts       []string `json:"hosts,omitempty" yaml:"hosts" bson:"hosts"`
	Serial      string   `json:"serial,omitzero" yaml:"serial" bson:"serial"`
	Product     `json:"product" yaml:"product" bson:"product"`
	Issue       `json:"issue" yaml:"issue" bson:"issue"`
	Quantity    string `json:"quantity" yaml:"quantity" bson:"quantity"`
	SupportPlan string `json:"supportPlan" yaml:"supportPlan" bson:"supportPlan"`
	Expiry      `json:"expiry" yaml:"expiry" bson:"expiry"`
	Status      status.License `json:"status" yaml:"status" bson:"status"`
}

type Product struct {
	Name    string `json:"name" yaml:"name" bson:"name"`
	Feature string `json:"feature" yaml:"feature" bson:"feature"`
}

type Issue struct {
	By       string `json:"by" yaml:"by" bson:"by"`
	To       string `json:"to" yaml:"to" bson:"to"`
	Hardware string `json:"hardware" yaml:"hardware" bson:"hardware"`
	Date     string `json:"date" yaml:"date" bson:"date"`
}

type Expiry struct {
	Date string `json:"date" yaml:"date" bson:"date"`
	Days int    `json:"days" yaml:"days" bson:"days"`
}

type RawLicense struct {
	Name     string `json:"name" yaml:"name" bson:"name"`
	Type     string `json:"type" yaml:"type" bson:"type"`
	Hostname string `json:"hostname" yaml:"hostname" bson:"hostname"`
	Product  string `json:"product" yaml:"product" bson:"product"`
	Feature  string `json:"feature" yaml:"feature" bson:"feature"`
	Quantity string `json:"quantity" yaml:"quantity" bson:"quantity"`
	SLA      string `json:"sla" yaml:"sla" bson:"sla"`
	Serial   string `json:"serial" yaml:"serial" bson:"serial"`
	Check    int    `json:"check" yaml:"check" bson:"check"`
	IssueBy  string `json:"issueby" yaml:"issueby" bson:"issueby"`
	IssueTo  string `json:"issueto" yaml:"issueto" bson:"issueto"`
	Hardware string `json:"hardware" yaml:"hardware" bson:"hardware"`
	Expiry   string `json:"expiry" yaml:"expiry" bson:"expiry"`
	Date     string `json:"date" yaml:"date" bson:"date"`
	Days     int    `json:"days" yaml:"days" bson:"days"`
}

func (l *License) Key() string {
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s-%s-%s-%d",
		l.Type,
		l.Product.Name,
		l.Serial,
		l.Issue.By,
		l.Issue.To,
		l.Hardware,
		l.Expiry.Date,
		l.Expiry.Days,
	)
}

func (l *License) InitValidStatus() {
	l.Status = status.License{Current: status.Valid}
}

func (l *License) InitExpiredStatus() {
	l.Status = status.License{Current: status.Expired}
}

func (l *License) InitInvalidHardwareStatus() {
	l.Status = status.License{Current: "unmatched hardware"}
}

func (l *License) InitInvalidSignatureStatus() {
	l.Status = status.License{Current: "invalid signature"}
}

func (l *License) InitCompromisedStatus() {
	l.Status = status.License{Current: "system compromised"}
}

func (l *License) IsValid() bool {
	return l.Expiry.Date != ""
}

// note:
// in the current search lib(bleve), the algo is not able to detect the string if it include uppercase
// we've tried a few different init settings, but the result is not as expected as always
// currenlty, the only way we found is to convert all the string to lower case and inject to searcher
func (l *License) GenSearchableObject() License {
	l.Type = strings.ToLower(l.Type)
	l.Name = strings.ToLower(l.Name)
	l.Product.Name = strings.ToLower(l.Product.Name)
	l.Product.Feature = strings.ToLower(l.Product.Feature)
	l.Serial = strings.ToLower(l.Serial)
	l.SupportPlan = strings.ToLower(l.SupportPlan)
	l.Issue.By = strings.ToLower(l.Issue.By)
	l.Issue.To = strings.ToLower(l.Issue.To)
	l.Issue.Hardware = strings.ToLower(l.Issue.Hardware)
	for i := range l.Hosts {
		l.Hosts[i] = strings.ToLower(l.Hosts[i])
	}

	return *l
}

func (r *RawLicense) IsUnlicense() bool {
	return r.Date == ""
}

func InitLicenseSearchIndex() error {
	if licenseSearcher != nil {
		return nil
	}

	var err error
	mapping := bleve.NewIndexMapping()
	licenseSearcher, err = bleve.NewMemOnly(mapping)
	return err
}

func GetLicenseSearcher() bleve.Index {
	return licenseSearcher
}

func SetLicenses(licenses []License) {
	updateLicense.Lock()
	defer updateLicense.Unlock()
	license = licenses
}

func GetLicenses() []License {
	return license
}
