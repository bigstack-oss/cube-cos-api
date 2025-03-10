package trigger

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	Triggers         = "triggers"
	DB               = "triggers"
	Collection       = "triggers"
	ResponsePolicyV2 = "/etc/policies/alert_trigger/alert_resp2_0.yml"
)

var DefaultOptions = []Options{
	{
		Name:        "Administrative Level Notification",
		Description: `Configure how you are going to be notified for system events and host alerts, including levels "warning", "error", and "critical".`,
		Match:       `"severity" == 'W' OR "severity" == 'E' OR "severity" == 'C'`,
		Attributes: []Attribute{
			{
				Name:    "severity",
				Type:    "string",
				Value:   "W",
				Enabled: false,
			},
			{
				Name:    "severity",
				Type:    "string",
				Value:   "E",
				Enabled: false,
			},
			{
				Name:    "severity",
				Type:    "string",
				Value:   "C",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "DEV",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "CPU",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "DSK",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "MEM",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "NET",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "SRV",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "VRT",
				Enabled: false,
			},
		},
		Enabled: false,
	},
	{
		Name:        "Instance Level Notification",
		Description: `Configure how you are going to be notified for instance alerts, including levels "warning", and "critical".`,
		Match:       `"severity" == 'W' OR "severity" == 'C'`,
		Attributes: []Attribute{
			{
				Name:    "severity",
				Type:    "string",
				Value:   "W",
				Enabled: false,
			},
			{
				Name:    "severity",
				Type:    "string",
				Value:   "E",
				Enabled: false,
			},
			{
				Name:    "severity",
				Type:    "string",
				Value:   "C",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "DEV",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "CPU",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "DSK",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "MEM",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "NET",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "SRV",
				Enabled: false,
			},
			{
				Name:    "category",
				Type:    "string",
				Value:   "VRT",
				Enabled: false,
			},
		},
		Enabled: false,
	},
}

type Policy struct {
	Name     string    `json:"name" yaml:"name"`
	Version  string    `json:"version" yaml:"version"`
	Enabled  bool      `json:"enabled" yaml:"enabled"`
	Triggers []Options `json:"triggers" yaml:"triggers"`
}

func (p *Policy) UpdateOrAppendTrigger(trigger Options) {
	if !p.existingTuningUpdated(trigger) {
		p.AppendTrigger(trigger)
	}
}

func (p *Policy) existingTuningUpdated(trigger Options) bool {
	for i, existing := range p.Triggers {
		if existing.Name == trigger.Name {
			p.Triggers[i].Name = trigger.Name
			p.Triggers[i].Description = trigger.Description
			p.Triggers[i].Match = trigger.Match
			p.Triggers[i].Response = trigger.Response
			p.Triggers[i].Enabled = trigger.Enabled
			return true
		}
	}

	return false
}

func (p *Policy) AppendTrigger(trigger Options) {
	p.Triggers = append(p.Triggers, trigger)
}

type Options struct {
	Id          string      `json:"-" yaml:"-" bson:"id"`
	Name        string      `json:"name" yaml:"name" bson:"name"`
	Description string      `json:"description" yaml:"description"`
	Match       string      `json:"-" yaml:"match"`
	Attributes  []Attribute `json:"attributes" yaml:"-"`
	Response    `json:"response" yaml:"response"`
	Enabled     bool            `json:"enabled" yaml:"enabled"`
	Status      *status.Details `json:"-" yaml:"-" bson:"status"`
}

func (o *Options) InitResponse() {
	o.Response.Types = []string{}
	o.Response.Slacks = []slack.Channel{}
	o.Response.Emails = []email.Recipient{}
}

func (o *Options) HasEmailRecipients() bool {
	return len(o.Response.Emails) > 0
}

func (o *Options) HasSlackChannels() bool {
	return len(o.Response.Slacks) > 0
}

func (o *Options) GenTaskUpdate() Options {
	return Options{
		Id:     o.Id,
		Name:   o.Name,
		Status: o.Status,
	}
}

func (o *Options) SetError() {
	o.Status.Current = status.Error
}

func (o *Options) SetCompleted() {
	o.Status.Current = status.Completed
}

type Response struct {
	Types  []string          `json:"types" yaml:"-"`
	Slacks []slack.Channel   `json:"slacks" yaml:"slacks"`
	Emails []email.Recipient `json:"emails" yaml:"emails"`
}

type Attribute struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   any    `json:"value"`
	Enabled bool   `json:"enabled"`
}
