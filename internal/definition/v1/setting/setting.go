package setting

import (
	"reflect"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	DB            = "settings"
	ReqCollection = "requests"
	ReqTTL        = 3600
	PolicyV1      = "/etc/policies/alert_setting/alert_setting1_0.yml"
)

type Options struct {
	Type  string `json:"type" bson:"type"`
	Key   string `json:"key" bson:"key"`
	Value any    `json:"value" bson:"value"`

	TitlePrefix *TitlePrefix      `json:"titlePrefix,omitempty" bson:"titlePrefix,omitempty"`
	Sender      *email.Sender     `json:"sender,omitempty" bson:"sender,omitempty"`
	Recipient   *email.Recipient  `json:"recipient,omitempty" bson:"recipient,omitempty"`
	Slack       *slack.ApiChannel `json:"slack,omitempty" bson:"slack,omitempty"`

	Status status.Settings `json:"status" bson:"status"`
}

type CosAlert struct {
	TitlePrefix string `json:"titlePrefix" yaml:"titlePrefix"`
	Sender      `json:"sender" yaml:"sender"`
	Receiver    `json:"receiver" yaml:"receiver"`
}

type Sender struct {
	Email *email.CosSender `json:"email,omitempty" yaml:"email,omitempty"`
}

type ApiAlert struct {
	TitlePrefix `json:"titlePrefix" bson:"titlePrefix"`
	Email       email.Options `json:"email" bson:"email"`
	Slack       slack.Options `json:"slack" bson:"slack"`
}

type Receiver struct {
	Emails []email.Recipient  `json:"emails" yaml:"emails"`
	Slacks []slack.CosChannel `json:"slacks" yaml:"slacks"`
}

type TitlePrefix struct {
	Value  string          `json:"value" bson:"value"`
	Status status.Settings `json:"status" bson:"status"`
}

func (t *TitlePrefix) InitUpdateStatus() {
	t.Status = initUpdateStatus()
}

func (o *Options) InitCreateStatus() {
	o.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Created,
		IsUpdating: true,
	}
}

func (o *Options) InitUpdateStatus() {
	o.Status = initUpdateStatus()
}

func (o *Options) InitDeleteStatus() {
	o.Status = initDeleteStatus()
}

func (o *Options) SetError() {
	o.Status.Current = status.Error
	o.Status.IsUpdating = false
}

func (o *Options) SetCompleted() {
	o.Status.Current = status.Ok
	o.Status.IsUpdating = false
}

func (o *Options) GenTaskUpdate() Options {
	return Options{
		Type:   o.Type,
		Key:    o.Key,
		Status: o.Status,
	}
}

func (a *ApiAlert) InitOkStatus() {
	a.TitlePrefix.Status.InitOkStatus()
	a.Email.InitOkStatus()
	a.Slack.InitOkStatus()
}

func (e *CosAlert) HasSender(host string) bool {
	if e.Sender.Email == nil {
		return false
	}

	return e.Sender.Email.Host == host
}

func (e *CosAlert) HasRecipient(address string) bool {
	for _, recipient := range e.Receiver.Emails {
		if recipient.Address == address {
			return true
		}
	}

	return false
}

func (e *CosAlert) GetSlackUrlByName(name string) string {
	for _, slack := range e.Receiver.Slacks {
		if slack.Channel == name {
			return slack.URL
		}
	}

	return ""
}

func (e *CosAlert) HasSlackChannel(channel slack.CosChannel) bool {
	for _, slack := range e.Receiver.Slacks {
		if slack.Channel == channel.Channel {
			return true
		}

		if slack.URL == channel.URL {
			return true
		}
	}

	return false
}

func (e *CosAlert) IsRecipientEqual(recipient email.Recipient) bool {
	for _, email := range e.Receiver.Emails {
		if reflect.DeepEqual(email, recipient) {
			return true
		}
	}

	return false
}

func (e *CosAlert) ConvertToApiSchema() ApiAlert {
	senders := []email.Sender{}
	if e.Sender.Email.Host != "" {
		senders = append(senders, e.Sender.Email.ConvertToApiSchema())
	}

	return ApiAlert{
		TitlePrefix: TitlePrefix{
			Value: e.TitlePrefix,
		},
		Email: email.Options{
			Senders:    senders,
			Recipients: e.Receiver.Emails,
		},
		Slack: slack.Options{
			Channels: convertToApiChannels(e.Receiver.Slacks),
		},
	}
}

func convertToApiChannels(channels []slack.CosChannel) []slack.ApiChannel {
	apiChannels := []slack.ApiChannel{}

	for _, channel := range channels {
		apiChannels = append(
			apiChannels,
			slack.ApiChannel{
				Name:        channel.Channel,
				URL:         channel.URL,
				Description: channel.Description,
			},
		)
	}

	return apiChannels
}

func initUpdateStatus() status.Settings {
	return status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
	}
}

func initDeleteStatus() status.Settings {
	return status.Settings{
		Current:    status.Deleting,
		Desired:    status.Deleted,
		IsUpdating: true,
	}
}
