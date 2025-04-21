package settings

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/email"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/slack"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	emailv1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	slackv1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func sendTrialEmail(sender emailv1.Sender, recipient string) error {
	err := email.Send(
		sender.Address(),
		sender.UserAuth(),
		sender.From,
		[]string{recipient},
		[]byte("Subject: Test Email\n\nHi, \nThis is a test email to verify that our email delivery system is working correctly. If you received this message, it means the email route is functioning as expected. No further action is required.\nBest."),
	)
	if err != nil {
		log.Errorf("settings: failed to send trial email (%s)", err.Error())
		return fmt.Errorf(
			"failed to send trial email, please make sure the email sender setting is correct",
		)
	}

	return nil
}

func sendTrialSlackMessage(channel slackv1.CosChannel) error {
	h, err := slack.NewHelper()
	if err != nil {
		log.Errorf("settings: failed to create slack helper (%s)", err.Error())
		return err
	}

	return h.SendWebhookMsg(
		channel.URL,
		"A trial message from Cube COS",
	)
}

func setSenderAsVerified(sender emailv1.Sender) error {
	mongo := mongo.GetGlobalHelper()
	return mongo.UpdateOne(
		v1.SettingsDB(),
		emailv1.SenderCollection,
		bson.M{"host": sender.Host},
		bson.M{"$set": bson.M{"accessVerified": true}},
		options.Update().SetUpsert(true),
	)
}
