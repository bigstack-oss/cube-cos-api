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
)

type helper struct {
	c     *gin.Context
	mongo *mongo.Helper

	handler     string
	task        *setting.Options
	titlePrefix setting.TitlePrefix
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler, mongo: mongo.GetGlobalHelper()}
	switch handler {
	case "listSettings":
		return h, nil
	case "patchTitlePrefix":
		return h, h.parseTitlePrefixPatchReq()
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
	return &apiPoliy, nil
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
