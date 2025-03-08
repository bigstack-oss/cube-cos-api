package settings

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getAllSettings() (*v1.Setting, error) {
	titlePrefix, err := getTitlePrefix()
	if err != nil {
		log.Errorf("failed to get title prefix (%s)", err.Error())
		return nil, err
	}

	senders, err := v1.GetEmailSenders()
	if err != nil {
		log.Errorf("failed to get email sender (%s)", err.Error())
		return nil, err
	}

	recipients, err := v1.GetEmailRecipients()
	if err != nil {
		log.Errorf("failed to get email recipient (%s)", err.Error())
		return nil, err
	}

	channels, err := v1.GetSlackChannels()
	if err != nil {
		log.Errorf("failed to get slack channel (%s)", err.Error())
		return nil, err
	}

	return &v1.Setting{
		TitlePrefix: titlePrefix,
		Email: email.Options{
			Senders:    senders,
			Recipients: recipients,
		},
		Slack: slack.Options{
			Channels: channels,
		},
	}, nil
}

func upsertTitlePrefix(titlePrefix string) error {
	h := mongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.TitlePrefixCollection(),
		bson.M{"value": bson.M{"$ne": ""}},
		bson.M{"$set": bson.M{"value": titlePrefix}},
		options.Update().SetUpsert(true),
	)
}

func getTitlePrefix() (string, error) {
	h := mongo.GetGlobalHelper()
	cursor, err := h.GetQueryCursor(
		v1.SettingsDB(),
		v1.TitlePrefixCollection(),
		bson.M{},
	)
	if err != nil {
		log.Errorf("failed to get cursor for email sender (%s)", err.Error())
		return "", err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer cursor.Close(ctx)
	return parseTitlePrefix(cursor)
}

func insertEmailSender(sender email.Sender) error {
	h := mongo.GetGlobalHelper()
	return h.Insert(
		v1.SettingsDB(),
		email.SenderCollection,
		sender,
	)
}

func updateEmailSender(sender email.Sender) error {
	h := mongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		email.SenderCollection,
		bson.M{"host": sender.Host},
		bson.M{
			"$set": bson.M{
				"host":           sender.Host,
				"port":           sender.Port,
				"username":       sender.Username,
				"password":       sender.Password,
				"email":          sender.Email,
				"accessVerified": sender.AccessVerified,
			},
		},
	)
}

func removeEmailSender(host string) error {
	h := mongo.GetGlobalHelper()
	return h.DeleteOne(
		v1.SettingsDB(),
		email.SenderCollection,
		bson.M{"host": host},
	)
}

func insertEmailRecipient(recipient email.Recipient) error {
	h := mongo.GetGlobalHelper()
	return h.Insert(
		v1.SettingsDB(),
		email.RecipientCollection,
		recipient,
	)
}

func updateEmailRecipient(recipient email.Recipient) error {
	h := mongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		email.RecipientCollection,
		bson.M{"email": recipient.Email},
		bson.M{
			"$set": bson.M{
				"note": recipient.Note,
			},
		},
	)
}

func removeEmailRecipient(recipient string) error {
	h := mongo.GetGlobalHelper()
	return h.DeleteOne(
		v1.SettingsDB(),
		email.RecipientCollection,
		bson.M{"email": recipient},
	)
}

func insertSlackChannel(channel slack.Channel) error {
	h := mongo.GetGlobalHelper()
	return h.Insert(
		v1.SettingsDB(),
		slack.ChannelCollection,
		channel,
	)
}

func getSlackChannel(name string) (*slack.Channel, error) {
	h := mongo.GetGlobalHelper()
	resp, err := h.Get(
		v1.SettingsDB(),
		slack.ChannelCollection,
		bson.M{"name": name},
	)
	if err != nil {
		log.Errorf("failed to get slack channel (%s)", err.Error())
		return nil, err
	}

	channel := slack.Channel{}
	err = resp.Decode(&channel)
	if err != nil {
		log.Errorf("failed to decode slack channel (%s)", err.Error())
		return nil, err
	}

	return &channel, nil
}

func updateSlackChannel(channel slack.Channel) error {
	h := mongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		slack.ChannelCollection,
		bson.M{"name": channel.Name},
		bson.M{
			"$set": bson.M{
				"channel":     channel.Name,
				"url":         channel.URL,
				"description": channel.Description,
			},
		},
	)
}

func removeSlackChannel(name string) error {
	h := mongo.GetGlobalHelper()
	return h.DeleteOne(
		v1.SettingsDB(),
		slack.ChannelCollection,
		bson.M{"name": name},
	)
}

func isSenderExist(sender string) bool {
	h := mongo.GetGlobalHelper()
	count, err := h.GetCount(
		v1.SettingsDB(),
		email.SenderCollection,
		bson.M{"host": sender},
	)
	if err != nil {
		log.Errorf("failed to get count of email sender (%s)", err.Error())
		return false
	}

	return count > 0
}

func isRecipientExist(recipient string) bool {
	h := mongo.GetGlobalHelper()
	count, err := h.GetCount(
		v1.SettingsDB(),
		email.RecipientCollection,
		bson.M{"email": recipient},
	)
	if err != nil {
		log.Errorf("failed to get count of email recipient (%s)", err.Error())
		return false
	}

	return count > 0
}

func isChannelExist(channel string) bool {
	h := mongo.GetGlobalHelper()
	count, err := h.GetCount(
		v1.SettingsDB(),
		slack.ChannelCollection,
		bson.M{"name": channel},
	)
	if err != nil {
		log.Errorf("failed to get count of slack channel (%s)", err.Error())
		return false
	}

	return count > 0
}
