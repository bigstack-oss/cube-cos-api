package settings

import (
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

func (a *Api) SetOkStatus() {
	a.TitlePrefix.Status.SetOk()
	a.Email.SetOk()
	a.Slack.SetOk()
}

type Api struct {
	TitlePrefix `json:"titlePrefix" bson:"titlePrefix"`
	Email       email.Options `json:"email" bson:"email"`
	Slack       slack.Options `json:"slack" bson:"slack"`
}

type Sender struct {
	Email *email.CosSender `json:"email,omitempty" yaml:"email,omitempty"`
}

type Receiver struct {
	Emails []email.Recipient  `json:"emails" yaml:"emails"`
	Slacks []slack.CosChannel `json:"slacks" yaml:"slacks"`
}

func (t *TitlePrefix) SetUpdating() {
	t.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
	}
}
