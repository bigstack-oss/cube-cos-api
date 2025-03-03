package settings

import (
	"net/smtp"

	"github.com/slack-go/slack"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

func sendTrialEmail(sender email.Sender, recipient string) error {
	return smtp.SendMail(
		sender.Address(),
		sender.UserAuth(),
		sender.Email,
		[]string{recipient},
		[]byte("Subject: Trial Email\n\nThis is a trial email from Cube COS."),
	)
}

func sendTrialSlackMessage(channel v1.Channel) error {
	api := slack.New(channel.URL)
	_, _, err := api.PostMessage(
		channel.Name,
		slack.MsgOptionText("Hello from CubeCOS", false),
	)
	return err
}
