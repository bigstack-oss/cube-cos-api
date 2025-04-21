package settings

import (
	"context"

	cubemongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type helper struct {
	c     *gin.Context
	mongo *cubemongo.Helper

	handler string
	task    *setting.Options
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler, mongo: cubemongo.GetGlobalHelper()}

	switch handler {
	case "listSettings":
		return h, nil
	case "updateTitlePrefix":
		return h, h.initTitlePrefixUpdateParams()
	case "createEmailSender":
		return h, h.initEmailSenderCreateParams()
	case "patchEmailSender":
		return h, h.initEmailSenderPatchParams()
	case "deleteEmailSender":
		return h, h.initEmailSenderDeleteParams()
	case "createEmailRecipient":
		return h, h.initEmailRecipientCreateParams()
	case "patchEmailRecipient":
		return h, h.initEmailRecipientPatchParams()
	case "deleteEmailRecipient":
		return h, h.initEmailRecipientDeleteParams()
	case "createSlackChannel":
		return h, h.initSlackChannelCreateParams()
	case "putSlackChannel":
		return h, h.initSlackChannelPatchParams()
	case "deleteSlackChannel":
		return h, h.initSlackChannelDeleteParams()
	}

	return h, nil
}

func (h *helper) listSettings() (*setting.ApiAlert, error) {
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Infof("settings(%s): failed to get settings: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	apiAlert := setting.ConvertToApiSchema()
	h.syncUpdateStatus(&apiAlert)
	h.eraseSenderPassword(&apiAlert.Email.Senders)
	h.syncEmailSenderVerification(&apiAlert.Email.Senders)

	return &apiAlert, nil
}

func (h *helper) syncEmailSenderVerification(senders *[]email.Sender) {
	for i, sender := range *senders {
		if h.isEmailSenderVerified(&sender) {
			(*senders)[i].AccessVerified = true
		}
	}
}

func (h *helper) isEmailSenderVerified(sender *email.Sender) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		email.SenderCollection,
		bson.M{"host": sender.Host, "accessVerified": true},
	)
	if err != nil {
		log.Errorf("settings: failed to check email sender verification (%s)", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) syncUpdateStatus(alert *setting.ApiAlert) {
	alert.InitOkStatus()
	h.syncTitlePrefixUpdate(&alert.TitlePrefix)
	h.syncEmailSenderUpdate(&alert.Email.Senders)
	h.syncEmailRecipientUpdate(&alert.Email.Recipients)
	h.syncSlackUpdate(&alert.Slack)
}

func (h *helper) syncTitlePrefixUpdate(titlePrefix *setting.TitlePrefix) {
	if h.isTitlePrefixUpdating() {
		h.syncTitlePrefixUpdateValue(titlePrefix)
		titlePrefix.InitUpdateStatus()
	}
}

func (h *helper) syncEmailSenderUpdate(senders *[]email.Sender) {
	if len(*senders) == 0 {
		h.syncUpdatingEmailSender(senders)
	}

	for i, sender := range *senders {
		if h.isEmailSenderUpdating(&sender) {
			(*senders)[i] = *h.syncEmailSenderUpdateValue(&sender)
			(*senders)[i].InitUpdateStatus()
		}
	}
}

func (h *helper) syncUpdatingEmailSender(senders *[]email.Sender) {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "emailSender"},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email sender cursor (%s)", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.parseUpdatingEmailSenders(c, senders)
}

func (h *helper) parseUpdatingEmailSenders(c *mongo.Cursor, senders *[]email.Sender) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &setting.Options{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode email sender cursor (%s)", err.Error())
			continue
		}

		(*senders) = append(*senders, *req.Sender)
	}
}

func (h *helper) syncEmailSenderUpdateValue(sender *email.Sender) *email.Sender {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "emailSender", "key": sender.Host},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email sender cursor (%s)", err.Error())
		return nil
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return h.parseEmailSenderUpdateValue(c)
}

func (h *helper) parseEmailSenderUpdateValue(c *mongo.Cursor) *email.Sender {
	req := &setting.Options{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode email sender cursor (%s)", err.Error())
			return nil
		}
	}

	return req.Sender
}

func (h *helper) syncEmailRecipientUpdate(recipients *[]email.Recipient) {
	h.syncUpdatingEmailRecipient(recipients)

	for i, recipient := range *recipients {
		if h.isEmailRecipientUpdating(&recipient) {
			(*recipients)[i] = *h.syncEmailRecipientUpdateValue(&recipient)
			(*recipients)[i].InitUpdateStatus()
		}
	}
}

func (h *helper) syncUpdatingEmailRecipient(recipients *[]email.Recipient) {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "emailRecipient"},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email recipient cursor (%s)", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.parseUpdatingEmailRecipients(c, recipients)

	uniqueRecipients := make(map[string]email.Recipient)
	for _, recipient := range *recipients {
		uniqueRecipients[recipient.Address] = recipient
	}

	*recipients = make([]email.Recipient, 0, len(uniqueRecipients))
	for _, recipient := range uniqueRecipients {
		*recipients = append(*recipients, recipient)
	}
}

func (h *helper) parseUpdatingEmailRecipients(c *mongo.Cursor, recipients *[]email.Recipient) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &setting.Options{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode email recipient cursor (%s)", err.Error())
			continue
		}

		*recipients = append(*recipients, *req.Recipient)
	}
}

func (h *helper) syncEmailRecipientUpdateValue(recipient *email.Recipient) *email.Recipient {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "emailRecipient", "recipient.address": recipient.Address},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email recipient cursor (%s)", err.Error())
		return nil
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return h.parseEmailRecipientUpdateValue(c)
}

func (h *helper) parseEmailRecipientUpdateValue(c *mongo.Cursor) *email.Recipient {
	req := &setting.Options{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode email recipient cursor (%s)", err.Error())
			return nil
		}
	}

	return req.Recipient
}

func (h *helper) syncSlackUpdate(slackOpts *slack.Options) {
	if len(slackOpts.Channels) == 0 {
		slackOpts.Channels = []slack.ApiChannel{}
	}

	h.syncUpdateSlackChannels(&slackOpts.Channels)

	for i, channel := range slackOpts.Channels {
		if h.isSlackUpdating(&channel) {
			slackOpts.Channels[i] = *h.syncSlackUpdateValue(&channel)
			slackOpts.Channels[i].InitUpdateStatus()
		}
	}
}

func (h *helper) syncSlackUpdateValue(channel *slack.ApiChannel) *slack.ApiChannel {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "slackChannel", "key": channel.Name},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from slack channel cursor (%s)", err.Error())
		return nil
	}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return h.parseSlackUpdateValue(c)
}

func (h *helper) parseSlackUpdateValue(c *mongo.Cursor) *slack.ApiChannel {
	req := &setting.Options{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode slack channel cursor (%s)", err.Error())
			return nil
		}
	}

	return req.Slack
}

func (h *helper) syncUpdateSlackChannels(channels *[]slack.ApiChannel) {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "slackChannel"},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from slack channel cursor (%s)", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.parseUpdatingSlackChannels(c, channels)

	uniqueChannels := make(map[string]slack.ApiChannel)
	for _, channel := range *channels {
		uniqueChannels[channel.URL] = channel
	}

	*channels = make([]slack.ApiChannel, 0, len(uniqueChannels))
	for _, channel := range uniqueChannels {
		*channels = append(*channels, channel)
	}
}

func (h *helper) parseUpdatingSlackChannels(c *mongo.Cursor, slack *[]slack.ApiChannel) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &setting.Options{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode slack channel cursor (%s)", err.Error())
			continue
		}

		*slack = append(*slack, *req.Slack)
	}
}

func (h *helper) syncTitlePrefixUpdateValue(titlePrefix *setting.TitlePrefix) {
	c, err := h.mongo.GetQueryCursor(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "titlePrefix"},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from title prefix cursor (%s)", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.parseTitlePrefixUpdateValue(c, titlePrefix)
}

func (h *helper) parseTitlePrefixUpdateValue(c *mongo.Cursor, titlePrefix *setting.TitlePrefix) {
	req := &setting.Options{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode title prefix cursor (%s)", err.Error())
			continue
		}

		titlePrefix.Value = req.Value.(string)
		break
	}
}
