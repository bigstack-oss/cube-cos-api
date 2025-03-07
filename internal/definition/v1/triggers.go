package v1

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

const (
	Triggers = "triggers"
)

var DefaultTriggers = []Trigger{
	{
		Name:        "Administrative Level Notification",
		Description: `Configure how you are going to be notified for system events and host alerts, including levels "warning", "error", and "critical".`,
		Match:       `"severity" == 'W' OR "severity" == 'E' OR "severity" == 'C'`,
		Attributes: []Attribute{
			{
				Name:  "severity",
				Type:  "string",
				Value: "W",
			},
			{
				Name:  "severity",
				Type:  "string",
				Value: "E",
			},
			{
				Name:  "severity",
				Type:  "string",
				Value: "C",
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
				Name:  "severity",
				Type:  "string",
				Value: "W",
			},
			{
				Name:  "severity",
				Type:  "string",
				Value: "C",
			},
		},
		Enabled: false,
	},
}

type Trigger struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	Match       string      `json:"match" yaml:"match"`
	Attributes  []Attribute `json:"attributes" yaml:"attributes"`
	Response    `json:"response" yaml:"response"`
	Enabled     bool `json:"enabled" yaml:"enabled"`
}

type Response struct {
	Types  []string          `json:"includes" yaml:"includes"`
	Slacks []slack.Channel   `json:"slacks" yaml:"slacks"`
	Emails []email.Recipient `json:"emails" yaml:"emails"`
}

type Attribute struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}
