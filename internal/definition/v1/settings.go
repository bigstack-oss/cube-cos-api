package v1

import (
	"context"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	Settings    = "settings"
	titlePrefix = "titlePrefix"
)

type Setting struct {
	TitlePrefix string        `json:"titlePrefix" bson:"titlePrefix"`
	Email       email.Options `json:"email" bson:"email"`
	Slack       slack.Options `json:"slack" bson:"slack"`
}

type TitlePrefix struct {
	Value string `json:"value" bson:"value"`
}

func GetSlackChannels() ([]slack.ApiChannel, error) {
	h := bsmongo.GetGlobalHelper()
	cursor, err := h.GetQueryCursor(
		Settings,
		slack.ChannelCollection,
		bson.M{},
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer cursor.Close(ctx)
	return parseChannels(cursor)
}

func parseChannels(cursor *mongo.Cursor) ([]slack.ApiChannel, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	channels := []slack.ApiChannel{}
	for cursor.Next(ctx) {
		channel := slack.ApiChannel{}
		err := cursor.Decode(&channel)
		if err != nil {
			continue
		}

		channels = append(channels, channel)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate slack channel(%s)", cursor.Err().Error())
	}

	return channels, nil
}

func TitlePrefixCollection() string {
	return titlePrefix
}

func SettingsDB() string {
	return Settings
}

func GetEmailRecipients() ([]email.Recipient, error) {
	h := bsmongo.GetGlobalHelper()
	c, err := h.GetQueryCursor(
		Settings,
		email.RecipientCollection,
		bson.M{},
	)
	if err != nil {
		log.Errorf("failed to get cursor for email recipient (%s)", err.Error())
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	recipients, err := parseRecipient(c)
	if err != nil {
		log.Errorf("failed to parse email recipient (%s)", err.Error())
		return nil, err
	}

	return recipients, nil
}

func parseRecipient(cursor *mongo.Cursor) ([]email.Recipient, error) {
	recipients := []email.Recipient{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	for cursor.Next(ctx) {
		recipient := email.Recipient{}
		err := cursor.Decode(&recipient)
		if err != nil {
			continue
		}

		recipients = append(recipients, recipient)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email recipient(%s)", cursor.Err().Error())
	}

	return recipients, nil
}
