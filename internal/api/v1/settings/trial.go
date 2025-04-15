package settings

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/email"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/slack"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	v1email "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	v1slack "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func sendTrialEmail(sender v1email.Sender, recipient string) error {
	err := email.Send(
		sender.Address(),
		sender.UserAuth(),
		sender.Email,
		[]string{recipient},
		[]byte("Subject: [Cube COS] A Trial Email From Settings\n\nThis is a trial email from Cube COS."),
	)
	if err != nil {
		log.Errorf("settings: failed to send trial email (%s)", err.Error())
		return fmt.Errorf(
			"failed to send trial email, please make sure the email sender setting is correct",
		)
	}

	return nil
}

func sendTrialSlackMessage(channel v1slack.Channel) error {
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

func setSenderAsVerified(sender v1email.Sender) error {
	mongo := mongo.GetGlobalHelper()
	return mongo.UpdateOne(
		v1.SettingsDB(),
		v1email.SenderCollection,
		bson.M{"host": sender.Host},
		bson.M{"$set": bson.M{"accessVerified": true}},
		options.Update().SetUpsert(true),
	)
}
