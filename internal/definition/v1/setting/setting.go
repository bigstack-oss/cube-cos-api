package setting

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	DB            = "settings"
	ReqCollection = "requests"
	PolicyV1      = "/etc/policies/alert_setting/alert_setting1_0.yml"
)

type EtcPolicy struct {
	Name        string        `json:"name" yaml:"name"`
	Version     float64       `json:"version" yaml:"version"`
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	TitlePrefix string        `json:"titlePrefix" yaml:"titlePrefix"`
	Sender      *email.Sender `json:"sender,omitempty" yaml:"sender,omitempty"`
	Receiver    `json:"receiver" yaml:"receiver"`
}

type Receiver struct {
	Emails []email.Recipient `json:"emails" yaml:"emails"`
	Slacks []slack.Channel   `json:"slacks" yaml:"slacks"`
}

type ApiPolicy struct {
	TitlePrefix `json:"titlePrefix" bson:"titlePrefix"`
	Email       email.Options `json:"email" bson:"email"`
	Slack       slack.Options `json:"slack" bson:"slack"`
}

type TitlePrefix struct {
	Value  string          `json:"value" bson:"value"`
	Status status.Settings `json:"status" bson:"status"`
}

func (a *ApiPolicy) InitOkStatus() {
	a.TitlePrefix.Status.InitOkStatus()
	a.Email.InitOkStatus()
	a.Slack.InitOkStatus()
}

func (t *TitlePrefix) InitUpdateStatus() {
	t.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}
