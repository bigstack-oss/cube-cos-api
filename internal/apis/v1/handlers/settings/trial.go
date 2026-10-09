package settings

import (
	"fmt"

	emailv2 "github.com/bigstack-oss/bigstack-dependency-go/pkg/email/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/slack"
	emailv1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) sendEmail(sender *emailv1.Sender, recipient string) error {
	err := sendTrialEmail(sender, recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to send trial email: %v", h.reqId, err)
		return fmt.Errorf(
			"failed to send trial email, please make sure the email sender setting is correct",
		)
	}

	return nil
}

func sendTrialEmail(sender *emailv1.Sender, recipient string) error {
	opts := append(sender.EmailOptions(), emailv2.RetryCount(0))
	helper, err := emailv2.NewHelper(opts...)
	if err != nil {
		return err
	}

	return helper.Send(emailv2.Message{
		To:      []string{recipient},
		Subject: "Test Email",
		Body:    "Hi, \nThis is a test email to verify that our email delivery system is working correctly. If you received this message, it means the email route is functioning as expected. No further action is required.\nBest.",
	})
}

func (h *helper) sendSlackMessage() error {
	slack, err := slack.NewHelper()
	if err != nil {
		return err
	}

	return slack.SendWebhookMsg(
		h.slackChannel,
		"A trial message from Cube COS",
	)
}

func (h *helper) setSenderAsVerified(sender emailv1.Sender) error {
	return h.mongo.UpdateOne(
		settings.DB,
		emailv1.SenderCollection,
		bson.M{"host": sender.Host},
		bson.M{"$set": bson.M{"accessVerified": true}},
		options.Update().SetUpsert(true),
	)
}
