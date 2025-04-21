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
	h.syncEmailRecipientUpdate(&alert.Email.Recipients)
	h.syncEmailSenderUpdate(&alert.Email.Senders)
	h.syncSlackUpdate(&alert.Slack)
}

func (h *helper) syncTitlePrefixUpdate(titlePrefix *setting.TitlePrefix) {
	if h.isTitlePrefixUpdating() {
		h.updateTitlePrefixValue(titlePrefix)
		titlePrefix.InitUpdateStatus()
	}
}

func (h *helper) updateTitlePrefixValue(titlePrefix *setting.TitlePrefix) {
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

		titlePrefix.Value = req.Value
		break
	}
}

func (h *helper) syncEmailRecipientUpdate(recipients *[]email.Recipient) {
	for i, recipient := range *recipients {
		if h.isEmailRecipientUpdating(&recipient) {
			(*recipients)[i].InitUpdateStatus()
		}
	}
}

func (h *helper) syncEmailSenderUpdate(senders *[]email.Sender) {
	for i, sender := range *senders {
		if h.isEmailSenderUpdating(&sender) {
			(*senders)[i].InitUpdateStatus()
		}
	}
}

func (h *helper) syncSlackUpdate(slack *slack.Options) {
	for i, channel := range slack.Channels {
		if h.isSlackUpdating(&channel) {
			slack.Channels[i].InitUpdateStatus()
		}
	}
}
