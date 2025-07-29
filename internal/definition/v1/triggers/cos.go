package triggers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/script"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	json "github.com/json-iterator/go"
)

type Trigger struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Topic       string `json:"topic"`
	Match       string `json:"match"`
	Description string `json:"description"`
	Responses   `json:"responses"`

	Emails []email.Recipient  `json:"emails"`
	Slacks []slack.CosChannel `json:"slacks"`
	Execs  script.Execs       `json:"execs"`
}

type Responses struct {
	Emails []string `json:"emails"`
	Slacks []string `json:"slacks"`
	Execs  `json:"execs"`
}

type Execs struct {
	Shells []string `json:"shells"`
}

func (t *Trigger) Bytes() ([]byte, error) {
	return json.Marshal(t)
}

// note:
// the regexp.MustCompile(`"([^"]+)"\s*==\s*"([^"]+)"`) is used to
// match the attribute pairs -> "field" == "value"
func (t *Trigger) ParseAttributes() Attribute {
	regex := regexp.MustCompile(`"([^"]+)"\s*==\s*"([^"]+)"`)
	matches := regex.FindAllStringSubmatch(t.Match, -1)
	attrs := Attribute{
		AlertTypes: []string{},
		EventIds:   []string{},
		Severities: []string{},
		Categories: []string{},
	}

	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		val := strings.TrimSpace(match[2])
		switch key {
		case "name()":
			attrs.AlertTypes = append(attrs.AlertTypes, val)
		case "key":
			attrs.EventIds = append(attrs.EventIds, val)
		case "severity":
			attrs.Severities = append(attrs.Severities, events.GetSeverityFullName(val))
		case "category":
			attrs.Categories = append(attrs.Categories, val)
		}
	}

	return attrs
}

func (t *Trigger) ConvertToApiEmails() []email.Recipient {
	emails := []email.Recipient{}
	for _, e := range t.Emails {
		emails = append(
			emails,
			email.Recipient{Address: e.Address, Enabled: true},
		)
	}

	return emails
}

func (t *Trigger) ConvertToApiSlacks() []slack.ApiChannel {
	slacks := []slack.ApiChannel{}
	for _, s := range t.Slacks {
		slacks = append(
			slacks,
			slack.ApiChannel{URL: s.URL, Enabled: true},
		)
	}

	return slacks
}

func (t *Trigger) GenMatchRule(req ReqOpts) string {
	andRules := []string{}

	if len(req.Attribute.AlertTypes) > 0 {
		andRules = append(andRules, t.GenOrRule("type", req.Attribute.AlertTypes))
	}

	if len(req.Attribute.EventIds) > 0 {
		andRules = append(andRules, t.GenOrRule("id", req.Attribute.EventIds))
	}

	if len(req.Attribute.Severities) > 0 {
		andRules = append(andRules, t.GenOrRule("severity", req.Attribute.Severities))
	}

	if len(req.Attribute.Categories) > 0 {
		andRules = append(andRules, t.GenOrRule("category", req.Attribute.Categories))
	}

	return strings.Join(andRules, " AND ")
}

func (t *Trigger) GenOrRule(key string, attrs []string) string {
	rule := []string{}
	for _, attr := range attrs {
		rule = append(
			rule,
			fmt.Sprintf(`%q == %q`, key, attr),
		)
	}

	orRule := strings.Join(rule, " OR ")
	return fmt.Sprintf(`(%s)`, orRule)
}
