package licenses

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module = "licenses"
	Dir    = "/etc/update"

	CubeCOS = "CubeCOS"
	CubeCMP = "CubeCMP"
	NA      = "N/A"
)

var (
	licenses = atomic.Pointer[[]License]{}
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

type Attachment struct {
	SerialNumber string `json:"serialNumber"`
	Hostname     string `json:"hostname"`
	Role         string `json:"role"`
	Product      string `json:"product"`
	Status       string `json:"status"`
}

type Verification struct {
	License     `json:"license" yaml:"license" bson:"license"`
	EffectNodes []Node `json:"effectNodes" yaml:"effectNodes" bson:"effectNodes"`
}

type Node struct {
	Name   string `json:"name" yaml:"name" bson:"name"`
	Role   string `json:"role" yaml:"role" bson:"role"`
	Expiry `json:"expiry" yaml:"expiry" bson:"expiry"`
	Status status.License `json:"status" yaml:"status" bson:"status"`
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

func (l *License) SetValid() {
	l.Status = status.License{Current: status.Valid}
}

func (l *License) SetExpired() {
	l.Status = status.License{Current: status.Expired}
}

func (l *License) InitInvalidHardware() {
	l.Status = status.License{Current: "unmatched hardware"}
}

func (l *License) InitInvalidSignature() {
	l.Status = status.License{Current: "invalid signature"}
}

func (l *License) SetCompromised() {
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
	l.Type = search.NormalizedKeyword(l.Type)
	l.Name = search.NormalizedKeyword(l.Name)
	l.Product.Name = search.NormalizedKeyword(l.Product.Name)
	l.Product.Feature = search.NormalizedKeyword(l.Product.Feature)
	l.Serial = search.NormalizedKeyword(l.Serial)
	l.SupportPlan = search.NormalizedKeyword(l.SupportPlan)
	l.Issue.By = search.NormalizedKeyword(l.Issue.By)
	l.Issue.To = search.NormalizedKeyword(l.Issue.To)
	l.Issue.Hardware = search.NormalizedKeyword(l.Issue.Hardware)
	for i := range l.Hosts {
		l.Hosts[i] = search.NormalizedKeyword(l.Hosts[i])
	}

	return *l
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

func IsNotInstalled(list []License) bool {
	return len(list) == 0
}

func List() []License {
	list := licenses.Load()
	if list == nil {
		return []License{}
	}

	return *list
}

func SetList(list []License) {
	licenses.Swap(&list)
}

func LowerProductsInPlace(products []string) {
	for i, product := range products {
		products[i] = strings.ToLower(product)
	}
}
