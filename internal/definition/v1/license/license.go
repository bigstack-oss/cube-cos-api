package license

import (
	"fmt"
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	Module = "licenses"
	Dir    = "/etc/update"
)

var (
	license       = []Options{}
	updateLicense sync.Mutex
)

type Raw struct {
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

type Options struct {
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

type Attachment struct {
	SerialNumber string `json:"serialNumber"`
	Hostname     string `json:"hostname"`
	Role         string `json:"role"`
	Product      string `json:"product"`
	Status       string `json:"status"`
}

type Verification struct {
	Options     `json:"license" yaml:"license" bson:"license"`
	EffectNodes []Node `json:"effectNodes" yaml:"effectNodes" bson:"effectNodes"`
}

type Node struct {
	Name   string `json:"name" yaml:"name" bson:"name"`
	Role   string `json:"role" yaml:"role" bson:"role"`
	Expiry `json:"expiry" yaml:"expiry" bson:"expiry"`
	Status status.License `json:"status" yaml:"status" bson:"status"`
}

func (o *Options) Key() string {
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s-%s-%s-%d",
		o.Type,
		o.Product.Name,
		o.Serial,
		o.Issue.By,
		o.Issue.To,
		o.Hardware,
		o.Expiry.Date,
		o.Expiry.Days,
	)
}

func (o *Options) InitValidStatus() {
	o.Status = status.License{Current: status.Valid}
}

func (o *Options) InitExpiredStatus() {
	o.Status = status.License{Current: status.Expired}
}

func (o *Options) InitInvalidHardwareStatus() {
	o.Status = status.License{Current: "unmatched hardware"}
}

func (o *Options) InitInvalidSignatureStatus() {
	o.Status = status.License{Current: "invalid signature"}
}

func (o *Options) InitCompromisedStatus() {
	o.Status = status.License{Current: "system compromised"}
}

func (o *Options) IsValid() bool {
	return o.Expiry.Date != ""
}

// note:
// in the current search lib(bleve), the algo is not able to detect the string if it include uppercase
// we've tried a few different init settings, but the result is not as expected as always
// currenlty, the only way we found is to convert all the string to lower case and inject to searcher
func (o *Options) GenSearchableObject() Options {
	o.Type = search.NormalizedKeyword(o.Type)
	o.Name = search.NormalizedKeyword(o.Name)
	o.Product.Name = search.NormalizedKeyword(o.Product.Name)
	o.Product.Feature = search.NormalizedKeyword(o.Product.Feature)
	o.Serial = search.NormalizedKeyword(o.Serial)
	o.SupportPlan = search.NormalizedKeyword(o.SupportPlan)
	o.Issue.By = search.NormalizedKeyword(o.Issue.By)
	o.Issue.To = search.NormalizedKeyword(o.Issue.To)
	o.Issue.Hardware = search.NormalizedKeyword(o.Issue.Hardware)
	for i := range o.Hosts {
		o.Hosts[i] = search.NormalizedKeyword(o.Hosts[i])
	}

	return *o
}

func (o *Attachment) GenSearchableObject() Attachment {
	return Attachment{
		SerialNumber: search.NormalizedKeyword(o.SerialNumber),
		Hostname:     search.NormalizedKeyword(o.Hostname),
		Role:         search.NormalizedKeyword(o.Role),
		Product:      search.NormalizedKeyword(o.Product),
		Status:       search.NormalizedKeyword(o.Status),
	}
}

func (r *Raw) IsUnlicense() bool {
	return r.Date == ""
}

func SetList(licenses []Options) {
	updateLicense.Lock()
	defer updateLicense.Unlock()
	license = licenses
}

func GetList() []Options {
	return license
}
