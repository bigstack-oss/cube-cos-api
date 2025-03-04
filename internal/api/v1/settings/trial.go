package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/email"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/slack"
	v1email "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	v1slack "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

func sendTrialEmail(sender v1email.Sender, recipient string) error {
	return email.Send(
		sender.Address(),
		sender.UserAuth(),
		sender.Email,
		[]string{recipient},
		[]byte("Subject: Trial Email\n\nThis is a trial email from Cube COS."),
	)
}

func sendTrialSlackMessage(channel v1slack.Channel) error {
	h, err := slack.NewHelper(slack.Token(channel.URL))
	if err != nil {
		return err
	}

	return h.SendTextMsg(
		channel.Name,
		"This is a trial message from Cube COS.",
	)
}
