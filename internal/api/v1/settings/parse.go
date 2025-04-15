package settings

import (
	"context"
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *helper) initTitlePrefixPatchParams() error {
	h.task = &setting.Options{Type: "titlePrefix"}
	err := h.c.ShouldBindJSON(&h.task.TitlePrefix)
	if err != nil {
		return err
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailSenderCreateParams() error {
	h.task = &setting.Options{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailSenderPatchParams() error {
	host := h.c.Param("senderHost")
	if host == "" {
		return errors.New("email sender host is empty")
	}

	h.task = &setting.Options{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		return err
	}

	if h.task.Sender.Host == "" {
		h.task.Sender.Host = host
	}

	h.task.Sender.ResetAccessVerification()
	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailSenderDeleteParams() error {
	host := h.c.Param("senderHost")
	if host == "" {
		return errors.New("email sender host is empty")
	}

	h.task = &setting.Options{Type: "emailSender"}
	h.task.Sender = &email.Sender{Host: host}
	h.task.InitDeleteStatus()
	return nil
}

func (h *helper) initEmailRecipientCreateParams() error {
	h.task = &setting.Options{Type: "emailRecipient"}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	err = email.CheckFormat(h.task.Recipient.Address)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %s", api.GetReqId(h.c), err.Error())
		api.SetBadRequest(h.c, err)
		return err
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailRecipientPatchParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.New("email recipient email is empty")
	}

	h.task = &setting.Options{Type: "emailRecipient", Key: recipientEmail}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		return err
	}

	if h.task.Recipient.Address == "" {
		h.task.Recipient.Address = recipientEmail
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailRecipientDeleteParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.New("email recipient email is empty")
	}

	h.task = &setting.Options{Type: "emailRecipient"}
	h.task.Recipient = &email.Recipient{Address: recipientEmail}
	h.task.InitDeleteStatus()
	return nil
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

// func parseEmailSenderUpdate(c *gin.Context) (*email.Sender, error) {
// 	sender := email.Sender{}
// 	err := c.ShouldBindJSON(&sender)
// 	if err != nil {
// 		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
// 		return nil, err
// 	}

// 	if sender.Host == "" {
// 		sender.Host = c.Param("senderHost")
// 	}

// 	sender.ResetAccessVerification()
// 	return &sender, nil
// }

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
