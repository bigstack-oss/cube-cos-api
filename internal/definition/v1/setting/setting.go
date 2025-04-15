package setting

import (
	"reflect"

	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	DB            = "settings"
	ReqCollection = "requests"
	PolicyV1      = "/etc/policies/alert_setting/alert_setting1_0.yml"
)

type Options struct {
	Type string `json:"type" bson:"type"`
	Key  string `json:"key" bson:"-"`

	TitlePrefix *TitlePrefix     `json:"titlePrefix,omitempty" bson:"titlePrefix,omitempty"`
	Sender      *email.Sender    `json:"sender,omitempty" bson:"sender,omitempty"`
	Recipient   *email.Recipient `json:"recipient,omitempty" bson:"recipient,omitempty"`
	Slack       *slack.Channel   `json:"slack,omitempty" bson:"slack,omitempty"`

	Status status.Settings `json:"status" bson:"status"`
}

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

func (t *TitlePrefix) InitUpdateStatus() {
	t.Status = initUpdateStatus()
}

func (o *Options) InitUpdateStatus() {
	o.Status = initUpdateStatus()
}

func (o *Options) InitDeleteStatus() {
	o.Status = initDeleteStatus()
}

func (o *Options) GetKey() string {
	key := ""

	switch o.Type {
	case "titlePrefix":
		key = o.Type
	case "emailSender":
		key = o.Sender.Host
	case "emailRecipient":
		key = o.Recipient.Address
	case "slack":
		key = o.Slack.Name
	}

	return key
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

func (a *ApiPolicy) InitOkStatus() {
	a.TitlePrefix.Status.InitOkStatus()
	a.Email.InitOkStatus()
	a.Slack.InitOkStatus()
}

func (e *EtcPolicy) HasSender(host string) bool {
	if e.Sender == nil {
		return false
	}

	return e.Sender.Host == host
}

func (e *EtcPolicy) HasRecipient(address string) bool {
	for _, recipient := range e.Receiver.Emails {
		if recipient.Address == address {
			return true
		}
	}

	return false
}

func (e *EtcPolicy) IsTitlePrefixEqual(titlePrefix string) bool {
	return e.TitlePrefix == titlePrefix
}

func (e *EtcPolicy) IsSenderEqual(sender email.Sender) bool {
	if e.Sender == nil {
		return false
	}

	return reflect.DeepEqual(*e.Sender, sender)
}

func (e *EtcPolicy) IsRecipientEqual(recipient email.Recipient) bool {
	for _, email := range e.Receiver.Emails {
		if reflect.DeepEqual(email, recipient) {
			return true
		}
	}

	return false
}

func (e *EtcPolicy) UpdateOrAppendSetting(setting Options) {
	if !e.existingSettingUpdated(setting) {
		e.AppendSetting(setting)
	}
}

func (e *EtcPolicy) DeleteSetting(setting Options) {
	switch setting.Type {
	case "emailSender":
		e.Sender = nil
	case "emailRecipient":
		e.deleteRecipient(setting)
	}
}

func (e *EtcPolicy) existingSettingUpdated(setting Options) bool {
	switch setting.Type {
	case "titlePrefix":
		e.TitlePrefix = setting.TitlePrefix.Value
		return true
	case "emailSender":
		e.Sender = setting.Sender
		return true
	case "emailRecipient":
		return e.updateEmailRecipient(setting)
	}

	return false
}

func (e *EtcPolicy) deleteRecipient(setting Options) {
	for i, recipient := range e.Receiver.Emails {
		if recipient.Address == setting.Recipient.Address {
			e.Receiver.Emails = slices.Delete(e.Receiver.Emails, i, i+1)
			return
		}
	}
}

func (e *EtcPolicy) updateEmailRecipient(setting Options) bool {
	for i, recipient := range e.Receiver.Emails {
		if recipient.Address == setting.Key {
			e.Receiver.Emails[i] = *setting.Recipient
			return true
		}
	}

	return false
}

func (e *EtcPolicy) AppendSetting(setting Options) {
	switch setting.Type {
	case "emailRecipient":
		e.Receiver.Emails = append(e.Receiver.Emails, *setting.Recipient)
	case "slack":
		e.Receiver.Slacks = append(e.Receiver.Slacks, *setting.Slack)
	}
}

func initUpdateStatus() status.Settings {
	return status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}

func initDeleteStatus() status.Settings {
	return status.Settings{
		Current:    status.Deleting,
		Desired:    status.Deleted,
		IsUpdating: true,
	}
}
