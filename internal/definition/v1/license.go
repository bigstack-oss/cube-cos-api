package v1

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	Licenses = "licenses"
)

type License struct {
	Type                  string   `json:"type" yaml:"type" bson:"type"`
	Hostname              string   `json:"hostname" yaml:"hostname" bson:"hostname"`
	Hosts                 []string `json:"hosts" yaml:"hosts" bson:"hosts"`
	Serial                string   `json:"serial" yaml:"serial" bson:"serial"`
	Product               `json:"product" yaml:"product" bson:"product"`
	Issue                 `json:"issue" yaml:"issue" bson:"issue"`
	Quantity              `json:"quantity" yaml:"quantity" bson:"quantity"`
	ServiceLevelAgreement `json:"serviceLevelAgreement" yaml:"sla" bson:"serviceLevelAgreement"`
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
	Value int    `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
}

type ServiceLevelAgreement struct {
	Uptime                 float32 `json:"uptime" yaml:"uptime" bson:"uptime"`
	Period                 string  `json:"period" yaml:"period" bson:"period"`
	MeanTimeBetweenFailure string  `json:"meanTimeBetweenFailure" yaml:"meanTimeBetweenFailure" bson:"meanTimeBetweenFailure"`
	MeanTimeToRecovery     string  `json:"meanTimeToRecovery" yaml:"meanTimeToRecovery" bson:"meanTimeToRecovery"`
}

type Expiry struct {
	Date string `json:"date" yaml:"date" bson:"date"`
	Days int    `json:"days" yaml:"days" bson:"days"`
}

type RawLicense struct {
	Type                  string `json:"type" yaml:"type" bson:"type"`
	Hostname              string `json:"hostname" yaml:"hostname" bson:"hostname"`
	Product               `json:"product" yaml:"product" bson:"product"`
	Quantity              `json:"quantity" yaml:"quantity" bson:"quantity"`
	ServiceLevelAgreement `json:"serviceLevelAgreement" yaml:"sla" bson:"serviceLevelAgreement"`
	Serial                string `json:"serial" yaml:"serial" bson:"serial"`
	Check                 int    `json:"check" yaml:"check" bson:"check"`
	IssueBy               string `json:"issueby" yaml:"issueby" bson:"issueby"`
	IssueTo               string `json:"issueto" yaml:"issueto" bson:"issueto"`
	Hardware              string `json:"hardware" yaml:"hardware" bson:"hardware"`
	Expiry                string `json:"expiry" yaml:"expiry" bson:"expiry"`
	Date                  string `json:"date" yaml:"date" bson:"date"`
	Days                  int    `json:"days" yaml:"days" bson:"days"`
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
