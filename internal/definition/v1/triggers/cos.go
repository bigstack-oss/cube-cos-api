package triggers

import (
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	json "github.com/json-iterator/go"
)

type CosSchema struct {
	Name           string             `json:"name"`
	Enabled        bool               `json:"enabled"`
	Topic          string             `json:"topic"`
	Match          string             `json:"match"`
	Description    string             `json:"description"`
	Emails         []email.Recipient  `json:"emails"`
	Slacks         []slack.CosChannel `json:"slacks"`
	WriteResponses `json:"responses"`
	Execs          `json:"execs"`
}

type WriteResponses struct {
	Emails []string `json:"emails"`
	Slacks []string `json:"slacks"`
	Execs  []string `json:"execs"`
}

type Execs struct {
	Shells []string `json:"shells"`
	Bins   []string `json:"bins"`
}

func (c *CosSchema) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CosSchema) ToApiSchema() ApiSchema {
	return ApiSchema{
		Name:        c.Name,
		Description: c.Description,
		Match:       c.Match,
		Attributes:  c.ConvertToApiAttributes(),
		Response: Response{
			Emails: c.ConvertToApiEmails(),
			Slacks: c.ConvertToApiSlacks(),
		},
		Enabled: c.Enabled,
	}
}

func (c *CosSchema) ConvertToApiAttributes() []Attribute {
	enabledAttrs := []Attribute{}
	matchRule := strings.ReplaceAll(c.Match, `"`, ``)
	parts := strings.SplitSeq(matchRule, " OR ")
	for part := range parts {
		attrPair := strings.Split(part, " == ")
		if !isValidAttrPair(attrPair) {
			continue
		}

		enabledAttrs = append(
			enabledAttrs,
			Attribute{
				Name:  strings.TrimSpace(attrPair[0]),
				Value: strings.TrimSpace(attrPair[1]),
			},
		)
	}

	return enabledAttrs
}

func (c *CosSchema) ConvertToApiEmails() []email.Recipient {
	emails := []email.Recipient{}
	for _, e := range c.Emails {
		emails = append(
			emails,
			email.Recipient{Address: e.Address, Enabled: true},
		)
	}

	return emails
}

func (c *CosSchema) ConvertToApiSlacks() []slack.ApiChannel {
	slacks := []slack.ApiChannel{}
	for _, s := range c.Slacks {
		slacks = append(
			slacks,
			slack.ApiChannel{URL: s.URL, Enabled: true},
		)
	}

	return slacks
}

func isValidAttrPair(attrPair []string) bool {
	return len(attrPair) == 2
}

func isBuiltInTrigger(name string) bool {
	_, found := builtInNameMap[name]
	return found
}
