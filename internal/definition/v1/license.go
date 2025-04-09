package v1

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
)

const (
	Licenses = "licenses"
)

var (
	licenseSearcher bleve.Index
)

type License struct {
	Name                  string   `json:"name" yaml:"name" bson:"name"`
	Type                  string   `json:"type" yaml:"type" bson:"type"`
	Hostname              string   `json:"hostname,omitzero" yaml:"hostname" bson:"hostname"`
	Hosts                 []string `json:"hosts,omitempty" yaml:"hosts" bson:"hosts"`
	Serial                string   `json:"serial" yaml:"serial" bson:"serial"`
	Product               `json:"product" yaml:"product" bson:"product"`
	Issue                 `json:"issue" yaml:"issue" bson:"issue"`
	Quantity              `json:"quantity" yaml:"quantity" bson:"quantity"`
	ServiceLevelAgreement string `json:"serviceLevelAgreement" yaml:"sla" bson:"serviceLevelAgreement"`
	Expiry                `json:"expiry" yaml:"expiry" bson:"expiry"`
	Status                status.License `json:"status" yaml:"status" bson:"status"`
}

type Product struct {
	Name     string   `json:"name" yaml:"name" bson:"name"`
	Features []string `json:"features" yaml:"features" bson:"features"`
}

type Issue struct {
	By       string `json:"by" yaml:"by" bson:"by"`
	To       string `json:"to" yaml:"to" bson:"to"`
	Hardware string `json:"hardware" yaml:"hardware" bson:"hardware"`
	Date     string `json:"date" yaml:"date" bson:"date"`
}

type Quantity struct {
	Type  string `json:"type" yaml:"type" bson:"type"`
	Value int    `json:"value" yaml:"vcpu" bson:"value"`
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
