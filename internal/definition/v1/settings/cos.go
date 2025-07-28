package settings

import (
	"reflect"
	"sync/atomic"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

var (
	cosSchema = atomic.Pointer[*Cos]{}
)

type Cos struct {
	TitlePrefix string `json:"titlePrefix" yaml:"titlePrefix"`
	Sender      `json:"sender" yaml:"sender"`
	Receiver    `json:"receiver" yaml:"receiver"`
}

func (c *Cos) HasSender(host string) bool {
	if c.Sender.Email == nil {
		return false
	}

	return c.Sender.Email.Host == host
}

func (c *Cos) HasRecipient(address string) bool {
	for _, recipient := range c.Receiver.Emails {
		if recipient.Address == address {
			return true
		}
	}

	return false
}

func (c *Cos) GetEmail(email string) (*email.Recipient, bool) {
	for _, recipient := range c.Receiver.Emails {
		if recipient.Address == email {
			return &recipient, true
		}
	}

	return nil, false
}

func (c *Cos) GetSlackUrlByName(name string) string {
	for _, slack := range c.Receiver.Slacks {
		if slack.Channel == name {
			return slack.URL
		}
	}

	return ""
}

func (c *Cos) GetSlack(url string) (*slack.CosChannel, bool) {
	for _, slack := range c.Receiver.Slacks {
		if slack.URL == url {
			return &slack, true
		}
	}

	return nil, false
}

func (c *Cos) HasSlack(channel slack.CosChannel) bool {
	for _, slack := range c.Receiver.Slacks {
		if slack.Channel == channel.Channel {
			return true
		}

		if slack.URL == channel.URL {
			return true
		}
	}

	return false
}

func (c *Cos) IsRecipientEqual(recipient email.Recipient) bool {
	for _, email := range c.Receiver.Emails {
		if reflect.DeepEqual(email, recipient) {
			return true
		}
	}

	return false
}

func (c *Cos) ToApiSchema() Api {
	senders := []email.Sender{}
	if c.Sender.Email.Host != "" {
		senders = append(senders, c.Sender.Email.ToApiSchema())
	}

	return Api{
		TitlePrefix: TitlePrefix{
			Value: c.TitlePrefix,
		},
		Email: email.Options{
			Senders:    senders,
			Recipients: c.Receiver.Emails,
		},
		Slack: slack.Options{
			Channels: convertToApiChannels(c.Receiver.Slacks),
		},
	}
}

func GetCosSchema() *Cos {
	schema := cosSchema.Load()
	if schema == nil {
		return &Cos{}
	}

	return *schema
}

func SetCosSchema(alert *Cos) {
	cosSchema.Swap(&alert)
}
