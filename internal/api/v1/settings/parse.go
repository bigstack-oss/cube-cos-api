package settings

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func parseTitlePrefix(cursor *mongo.Cursor) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	for cursor.Next(ctx) {
		titlePrefix := v1.TitlePrefix{}
		err := cursor.Decode(&titlePrefix)
		if err != nil {
			continue
		}

		return titlePrefix.Value, nil
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email sender(%s)", cursor.Err().Error())
	}

	return "", nil
}

func parseEmailSender(cursor *mongo.Cursor) ([]email.Sender, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	senders := []email.Sender{}
	for cursor.Next(ctx) {
		sender := email.Sender{}
		err := cursor.Decode(&sender)
		if err != nil {
			continue
		}

		senders = append(senders, sender)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email sender(%s)", cursor.Err().Error())
	}

	return senders, nil
}

func parseEmailRecipient(cursor *mongo.Cursor) ([]email.Recipient, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	recipients := []email.Recipient{}
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

func parseSlackChannel(cursor *mongo.Cursor) ([]slack.Channel, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	channels := []slack.Channel{}
	for cursor.Next(ctx) {
		channel := slack.Channel{}
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
