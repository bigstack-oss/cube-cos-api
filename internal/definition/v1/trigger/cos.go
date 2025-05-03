package trigger

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	json "github.com/json-iterator/go"
)

type CosOptions struct {
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	Topic       string   `json:"topic"`
	Match       string   `json:"match"`
	Description string   `json:"description"`
	Emails      []string `json:"emails"`
	Slacks      []string `json:"slacks"`
	Execs       `json:"execs"`
}

type Execs struct {
	Shells []string `json:"shells"`
	Bins   []string `json:"bins"`
}

func (c *CosOptions) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CosOptions) ConvertToApiOptions() ApiOptions {
	return ApiOptions{
		Name:        c.Name,
		Description: c.Description,
		Match:       c.Match,
		Response: Response{
			Emails: c.ConvertToApiEmails(),
			Slacks: c.ConvertToApiSlacks(),
		},
		Enabled: c.Enabled,
	}
}

func (c *CosOptions) ConvertToApiEmails() []email.Recipient {
	emails := []email.Recipient{}
	for _, e := range c.Emails {
		emails = append(
			emails,
			email.Recipient{Address: e, Enabled: true},
		)
	}

	return emails
}

func (c *CosOptions) ConvertToApiSlacks() []slack.ApiChannel {
	slacks := []slack.ApiChannel{}
	for _, s := range c.Slacks {
		slacks = append(
			slacks,
			slack.ApiChannel{URL: s, Enabled: true},
		)
	}

	return slacks
}
