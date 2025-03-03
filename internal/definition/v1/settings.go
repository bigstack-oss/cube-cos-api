package v1

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

const (
	Settings        = "settings"
	titlePrefix     = "titlePrefix"
	emailSenders    = "emailSenders"
	emailRecipients = "emailRecipients"
	slackChannels   = "slackChannels"
)

type Setting struct {
	TitlePrefix string        `json:"titlePrefix" bson:"titlePrefix"`
	Email       email.Options `json:"email" bson:"email"`
	Slack       slack.Options `json:"slack" bson:"slack"`
}

type TitlePrefix struct {
	Value string `json:"value" bson:"value"`
}

func TitlePrefixCollection() string {
	return titlePrefix
}

func EmailSenderCollection() string {
	return emailSenders
}

func EmailRecipientCollection() string {
	return emailRecipients
}

func SlackChannelCollection() string {
	return slackChannels
}

func SettingsDB() string {
	return Settings
}
