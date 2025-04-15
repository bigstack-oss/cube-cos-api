package settings

import (
	"context"
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *helper) parseTitlePrefixPatchReq() error {
	return h.c.ShouldBindJSON(&h.titlePrefix)
}

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

func parseEmailSenderUpdate(c *gin.Context) (*email.Sender, error) {
	sender := email.Sender{}
	err := c.ShouldBindJSON(&sender)
	if err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
		return nil, err
	}

	if sender.Host == "" {
		sender.Host = c.Param("senderHost")
	}

	sender.ResetAccessVerification()
	return &sender, nil
}

func parseSlackChannelUpdate(c *gin.Context) (*slack.Channel, error) {
	channel := slack.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		return nil, err
	}

	if channel.Name == "" {
		channel.Name = c.Param("channelName")
		return &channel, nil
	}

	return &channel, nil
}

func parseEmailRecipientUpdate(c *gin.Context, recipient *email.Recipient) error {
	if recipient.Address == "" {
		recipient.Address = c.Param("recipientEmail")
	}

	return nil
}

func (h *helper) parseTaskUpdate() error {
	err := h.c.ShouldBindJSON(&h.task)
	if err != nil {
		log.Errorf("request(%s): failed to parse task: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	if h.task.Type == "" {
		err := errors.New("task type is empty")
		log.Errorf("request(%s): %v", api.GetReqId(h.c), err)
		return err
	}

	if h.task.Key == "" {
		err := errors.New("task key is empty")
		log.Errorf("request(%s): %v", api.GetReqId(h.c), err)
		return err
	}

	return nil
}
