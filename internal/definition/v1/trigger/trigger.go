package trigger

import (
	"fmt"
	"strings"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/shirou/gopsutil/v4/host"
	log "go-micro.dev/v5/logger"
)

const (
	Triggers         = "triggers"
	DB               = "triggers"
	Collection       = "triggers"
	ReqCollection    = "requests"
	ReqTTL           = 3600
	ResponsePolicyV2 = "/etc/policies/alert_resp/alert_resp2_0.yml"
	ISO8601Z         = "2006-01-02T15:04:05+00:00"
)

var DefaultOptions = []Options{
	{
		Name:        "Administrative Level Notification",
		Description: `Configure how you are going to be notified for system events and host alerts, including levels 'warning', 'error', and 'critical'.`,
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
		Description: `Configure how you are going to be notified for instance alerts, including levels 'warning', and 'critical'.`,
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
	Version  float64   `json:"version" yaml:"version"`
	Enabled  bool      `json:"enabled" yaml:"enabled"`
	Triggers []Options `json:"triggers" yaml:"triggers"`
}

type Toggle struct {
	Enable bool `json:"enable" yaml:"enable"`
}

func (p *Policy) GetTrigger(name string) Options {
	for _, trigger := range p.Triggers {
		if trigger.Name == name {
			return trigger
		}
	}

	return Options{}
}

func (p *Policy) UpdateOrAppendTrigger(trigger Options) {
	if !p.existingTriggerUpdated(trigger) {
		p.AppendTrigger(trigger)
	}
}

func (p *Policy) existingTriggerUpdated(trigger Options) bool {
	for i, existing := range p.Triggers {
		if existing.Name == trigger.Name {
			p.Triggers[i].Name = trigger.Name
			p.Triggers[i].Description = trigger.Description
			p.Triggers[i].Match = trigger.GenMatchRule()
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
	Name        string      `json:"name" yaml:"name" bson:"name"`
	Description string      `json:"description" yaml:"description"`
	Match       string      `json:"-" yaml:"match"`
	Attributes  []Attribute `json:"attributes" bson:"-" yaml:"-"`
	Response    `json:"response" yaml:"response"`
	Enabled     bool            `json:"enabled" yaml:"enabled"`
	Status      *status.Trigger `json:"status" yaml:"-" bson:"status"`
}

func (o *Options) InitResponse() {
	o.Response.Types = []string{}
	o.Response.Slacks = []slack.ApiChannel{}
	o.Response.Emails = []email.Recipient{}
}

func (o *Options) InitOkStatus() {
	o.Status = &status.Trigger{
		Current:    status.Ok,
		IsUpdating: false,
	}

	bootDuration, err := host.BootTime()
	if err != nil {
		o.Status.UpdatedAt = TimeISO8601Z(time.Now())
		return
	}

	bootTime := time.Unix(int64(bootDuration), 0)
	o.Status.UpdatedAt = TimeISO8601Z(bootTime)
}

func (o *Options) InitUpdateStatus() {
	o.Status = &status.Trigger{
		Current:    status.Updating,
		Desired:    status.Updated,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
		UpdatedAt:  time.Now().Local().Format(time.RFC3339),
		IsUpdating: true,
	}
}

func (o *Options) GenMatchRule() string {
	rule := []string{}
	for _, attr := range o.Attributes {
		if !attr.Enabled {
			continue
		}

		rule = append(
			rule,
			fmt.Sprintf(`%q == %q`, attr.Name, attr.Value),
		)
	}

	return strings.Join(rule, " OR ")
}

func (o *Options) HasEmailRecipients() bool {
	return len(o.Response.Emails) > 0
}

func (o *Options) HasSlackChannels() bool {
	return len(o.Response.Slacks) > 0
}

func (o *Options) GenTaskUpdate() Options {
	return Options{
		Name:   o.Name,
		Status: o.Status,
	}
}

func (o *Options) IsSame(trigger Options) bool {
	if o.Name != trigger.Name {
		log.Errorf("trigger name not same: %s != %s", o.Name, trigger.Name)
		return false
	}

	if o.Match != trigger.Match {
		log.Errorf("trigger match not same: %s != %s", o.Match, trigger.Match)
		return false
	}

	if o.Enabled != trigger.Enabled {
		log.Errorf("trigger enable not same: %v != %v", o.Enabled, trigger.Enabled)
		return false
	}

	// have to add a comparsion for response data
	return true
}

func (o *Options) SetError() {
	o.Status.Current = status.Error
}

func (o *Options) SetCompleted() {
	o.Status.Current = status.Ok
	o.Status.IsUpdating = false
}

type Response struct {
	Types  []string           `json:"types" yaml:"-"`
	Slacks []slack.ApiChannel `json:"slacks" yaml:"slacks"`
	Emails []email.Recipient  `json:"emails" yaml:"emails"`
}

type Attribute struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   any    `json:"value"`
	Enabled bool   `json:"enabled"`
}

func Get(name string) (*Options, bool) {
	for _, trigger := range DefaultOptions {
		if trigger.Name == name {
			return &trigger, true
		}
	}

	return nil, false
}

func TimeISO8601Z(t time.Time) string {
	return t.Format(ISO8601Z)
}
