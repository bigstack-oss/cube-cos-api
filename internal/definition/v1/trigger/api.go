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

var (
	List = []ApiOptions{}

	DefaultOptions = []ApiOptions{
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
)

type Toggle struct {
	Enable bool `json:"enable" yaml:"enable"`
}

type ApiOptions struct {
	Name        string      `json:"name" yaml:"name" bson:"name"`
	Description string      `json:"description" yaml:"description"`
	Match       string      `json:"-" yaml:"match"`
	Attributes  []Attribute `json:"attributes" bson:"-" yaml:"-"`
	Response    `json:"response" yaml:"response"`
	Enabled     bool            `json:"enabled" yaml:"enabled"`
	Status      *status.Trigger `json:"status" yaml:"-" bson:"status"`
}

func (a *ApiOptions) ConvertToCosOptions() CosOptions {
	return CosOptions{
		Name:        a.Name,
		Description: a.Description,
		Match:       a.GenMatchRule(),
		Emails:      a.GenEmailList(),
		Slacks:      a.GenSlackList(),
		Enabled:     a.Enabled,
	}
}

func (a *ApiOptions) GenEmailList() []string {
	emails := []string{}
	for _, email := range a.Response.Emails {
		if email.Enabled {
			emails = append(emails, email.Address)
		}
	}

	return emails
}

func (a *ApiOptions) GenSlackList() []string {
	slacks := []string{}
	for _, slack := range a.Response.Slacks {
		if slack.Enabled {
			slacks = append(slacks, slack.URL)
		}
	}

	return slacks
}

func (a *ApiOptions) InitOkStatus() {
	a.Status = &status.Trigger{
		Current:    status.Ok,
		IsUpdating: false,
	}

	bootDuration, err := host.BootTime()
	if err != nil {
		a.Status.UpdatedAt = TimeISO8601Z(time.Now())
		return
	}

	bootTime := time.Unix(int64(bootDuration), 0)
	a.Status.UpdatedAt = TimeISO8601Z(bootTime)
}

func (a *ApiOptions) InitUpdateStatus() {
	a.Status = &status.Trigger{
		Current:    status.Updating,
		Desired:    status.Updated,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
		UpdatedAt:  time.Now().Local().Format(time.RFC3339),
		IsUpdating: true,
	}
}

func (a *ApiOptions) GenMatchRule() string {
	rule := []string{}
	for _, attr := range a.Attributes {
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

func (a *ApiOptions) HasEmailRecipients() bool {
	return len(a.Response.Emails) > 0
}

func (a *ApiOptions) HasEmail(email string) bool {
	for _, recipient := range a.Response.Emails {
		if recipient.Address == email {
			return true
		}
	}

	return false
}

func (a *ApiOptions) SetEmailDetails(email email.Recipient) {
	for i, recipient := range a.Response.Emails {
		if recipient.Address == email.Address {
			a.Response.Emails[i].Enabled = email.Enabled
			a.Response.Emails[i].Note = email.Note
			return
		}
	}
}

func (a *ApiOptions) AppendEmail(email email.Recipient) {
	a.Response.Emails = append(a.Response.Emails, email)
}

func (a *ApiOptions) HasSlackChannels() bool {
	return len(a.Response.Slacks) > 0
}

func (a *ApiOptions) HasSlack(channel string) bool {
	for _, slack := range a.Response.Slacks {
		if slack.URL == channel {
			return true
		}
	}

	return false
}

func (a *ApiOptions) SetSlackDetails(slack slack.ApiChannel) {
	for i, channel := range a.Response.Slacks {
		if channel.URL == slack.URL {
			a.Response.Slacks[i].Name = slack.Name
			a.Response.Slacks[i].URL = slack.URL
			a.Response.Slacks[i].Description = slack.Description
			a.Response.Slacks[i].Enabled = slack.Enabled
			return
		}
	}
}

func (a *ApiOptions) AppendSlack(slack slack.ApiChannel) {
	a.Response.Slacks = append(a.Response.Slacks, slack)
}

func (a *ApiOptions) GenTaskUpdate() ApiOptions {
	return ApiOptions{
		Name:   a.Name,
		Status: a.Status,
	}
}

func (a *ApiOptions) IsSame(trigger ApiOptions) bool {
	if a.Name != trigger.Name {
		log.Errorf("trigger name not same: %s != %s", a.Name, trigger.Name)
		return false
	}

	if a.Match != trigger.Match {
		log.Errorf("trigger match not same: %s != %s", a.Match, trigger.Match)
		return false
	}

	if a.Enabled != trigger.Enabled {
		log.Errorf("trigger enable not same: %v != %v", a.Enabled, trigger.Enabled)
		return false
	}

	return true
}

func (a *ApiOptions) SetError() {
	a.Status.Current = status.Error
}

func (a *ApiOptions) SetCompleted() {
	a.Status.Current = status.Ok
	a.Status.IsUpdating = false
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

func Get(name string) (*ApiOptions, bool) {
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
