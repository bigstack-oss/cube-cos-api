package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type helper struct {
	c     *gin.Context
	mongo *mongo.Helper

	handler string
	task    *setting.Options
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler, mongo: mongo.GetGlobalHelper()}
	switch handler {
	case "listSettings":
		return h, nil
	case "patchTitlePrefix":
		return h, h.initTitlePrefixPatchParams()
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
	}

	return h, nil
}

func (h *helper) listSettings() (*setting.ApiPolicy, error) {
	etcPolicy, err := cubecos.GetEtcSettingPolicy()
	if err != nil {
		log.Infof("settings(%s): failed to get settings: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	apiPoliy := convertEtcPolicyToApiPolicy(etcPolicy)
	h.syncUpdateStatus(&apiPoliy)
	h.eraseSenderPassword(&apiPoliy.Email.Senders)
	h.syncEmailSenderVerification(&apiPoliy.Email.Senders)
	return &apiPoliy, nil
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

func (h *helper) syncUpdateStatus(apiPolicy *setting.ApiPolicy) {
	apiPolicy.InitOkStatus()
	h.syncTitlePrefixUpdate(&apiPolicy.TitlePrefix)
	h.syncEmailRecipientUpdate(&apiPolicy.Email.Recipients)
	h.syncEmailSenderUpdate(&apiPolicy.Email.Senders)
	h.syncSlackUpdate(&apiPolicy.Slack)
}

func (h *helper) syncTitlePrefixUpdate(titlePrefix *setting.TitlePrefix) {
	if h.isTitlePrefixUpdating() {
		titlePrefix.InitUpdateStatus()
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
