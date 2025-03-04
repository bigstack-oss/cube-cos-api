package settings

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/email"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/slack"
	v1email "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	v1slack "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"go-micro.dev/v5/logger"
)

func sendTrialEmail(sender v1email.Sender, recipient string) error {
	err := email.Send(
		sender.Address(),
		sender.UserAuth(),
		sender.Email,
		[]string{recipient},
		[]byte("Subject: Trial Email\n\nThis is a trial email from Cube COS."),
	)
	if err != nil {
		logger.Errorf("settings: failed to send trial email (%s)", err.Error())
		return fmt.Errorf(
			"failed to send trial email, please make sure the email sender setting is correct",
		)
	}

	return nil
}

func sendTrialSlackMessage(channel v1slack.Channel) error {
	h, err := slack.NewHelper()
	if err != nil {
		logger.Errorf("settings: failed to create slack helper (%s)", err.Error())
		return err
	}

	return h.SendWebhookMsg(
		channel.URL,
		"A trial message from Cube COS",
	)
}
