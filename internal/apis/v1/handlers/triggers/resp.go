package triggers

import (
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/shirou/gopsutil/v4/host"
)

var (
	builtInMap = map[string]triggerResp{
		"admin-notify": {
			Name:        "Administrative Level Notification",
			IsBuiltIn:   true,
			Description: `Configure how you are going to be notified for system events and host alerts, including levels 'warning', 'error', and 'critical'.`,
		},
		"instance-notify": {
			Name:        "Instance Level Notification",
			IsBuiltIn:   true,
			Description: `Configure how you are going to be notified for instance alerts, including levels 'warning', and 'critical'.`,
		},
	}

	builtInVariable = map[string]any{
		"id":            "alert id string",
		"message":       "alert message, from the TICKscript",
		"details":       "user-defined HTML content for a more detailed message",
		"time":          "YYYY-MM-DDTHH:MM:SSZ",
		"duration":      "an integer value in seconds",
		"level":         "OK or INFO or WARNING or CRITICAL",
		"previousLevel": "OK or INFO or WARNING or CRITICAL",
		"recoverable":   "bool value like true or false",
		"data": map[string]any{
			"time":  "YYYY-MM-DDTHH:MM:SSZ",
			"name":  "measurement name",
			"group": "group by tags concatenated",
			"tags": map[string]any{
				"tag_key_1": "tag value 1",
				"tag_key_2": "tag value 2",
			},
			"fields": map[string]any{
				"field_key_1": "some integer value",
				"field_key_2": "some string value",
			},
		},
	}
)

type triggerPage struct {
	Triggers   []triggerResp `json:"triggers"`
	pages.Page `json:"page"`
}

type triggerResp struct {
	Name        string `json:"name" yaml:"name" bson:"name"`
	IsBuiltIn   bool   `json:"isBuiltIn" yaml:"-" bson:"isBuiltIn"`
	Description string `json:"description" yaml:"description" bson:"description"`

	Topic              string `json:"topic" yaml:"topic" bson:"topic"`
	triggers.Attribute `json:"attribute" yaml:"attribute" bson:"attribute"`
	Response           `json:"response" yaml:"response" bson:"response"`

	Enabled bool            `json:"enabled" yaml:"enabled" bson:"enabled"`
	Status  *status.Trigger `json:"status" yaml:"-" bson:"status"`
}

type materials struct {
	triggers.Attribute `json:"attribute"`
	materialResp       `json:"response"`
}
type materialResp struct {
	ScriptType `json:"scriptType"`
	Emails     []Email `json:"emails"`
	Slacks     []Slack `json:"slacks"`
}

type Response struct {
	Types           []string `json:"types" yaml:"-" bson:"types"`
	triggers.Script `json:"script"`
	Emails          []Email `json:"emails"`
	Slacks          []Slack `json:"slacks"`
}

type ScriptType struct {
	Language        string `json:"language"`
	Environment     string `json:"environment"`
	BuiltInVariable `json:"builtInVariable"`
}

type BuiltInVariable struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Value       map[string]any `json:"value"`
}

type Notifications struct {
	Emails []Email `json:"emails"`
	Slacks []Slack `json:"slacks"`
}

type Email struct {
	Address string `json:"address"`
	Note    string `json:"note"`
}

type Slack struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

func (t *triggerResp) SetOk() {
	t.Status = &status.Trigger{
		Current:      status.Ok,
		IsProcessing: false,
	}

	bootDuration, err := host.BootTime()
	if err != nil {
		t.Status.UpdatedAt = time.TimeISO8601Z(ostime.Now())
		return
	}

	bootTime := ostime.Unix(int64(bootDuration), 0)
	t.Status.UpdatedAt = time.TimeISO8601Z(bootTime)
}

func (t *triggerResp) HasEmails() bool {
	return len(t.Response.Emails) > 0
}

func (t *triggerResp) HasEmail(email string) bool {
	for _, recipient := range t.Response.Emails {
		if recipient.Address == email {
			return true
		}
	}

	return false
}

func (t *triggerResp) HasSlacks() bool {
	return len(t.Response.Slacks) > 0
}

func (t *triggerResp) HasScript() bool {
	return t.Response.Script.Name != ""
}

func (h *helper) syncBuiltInInfo(resp *triggerResp) {
	details, found := builtInMap[resp.Name]
	if found {
		resp.Name = details.Name
		resp.IsBuiltIn = details.IsBuiltIn
		resp.Description = details.Description
	}
}

func (h *helper) syncResponseTypes(trigger *triggerResp) {
	trigger.Response.Types = []string{}
	if trigger.HasEmails() {
		trigger.Response.Types = append(trigger.Response.Types, "email")
	}

	if trigger.HasSlacks() {
		trigger.Response.Types = append(trigger.Response.Types, "slack")
	}

	if trigger.HasScript() {
		trigger.Response.Types = append(trigger.Response.Types, "script")
	}
}
