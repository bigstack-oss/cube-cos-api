package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) checkRecipientUpdate() error {
	if !h.isRecipientExist(h.c.Param("recipientEmail")) {
		return errors.New("recipient not found")
	}

	err := email.CheckFormat(h.c.Param("recipientEmail"))
	if err != nil {
		return errors.New("recipient email format is invalid")
	}

	return nil
}

func (h *helper) isRecipientExist(recipient string) bool {
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	return setting.HasRecipient(recipient)
}

func (h *helper) isSlackChannlExist() bool {
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	channel := h.task.Slack.ToCosSchema()
	if setting.HasSlack(channel) {
		return true
	}

	if h.hasUpdatingSlack() {
		return true
	}

	return false
}

func (h *helper) hasUpdatingSlack() bool {
	count, err := h.mongo.GetCount(
		settings.DB,
		settings.ReqCollection,
		bson.M{
			"type": "slackChannel",
			"$or": bson.A{
				bson.M{"key": h.c.Param("channelName")},
				bson.M{"slack.url": h.task.Slack.URL},
			},
		},
	)
	if err != nil {
		return true
	}

	return count > 0
}

func (h *helper) isSenderVerified(sender *email.Sender) bool {
	count, err := h.mongo.GetCount(
		settings.DB,
		email.SenderCollection,
		bson.M{"host": sender.Host, "accessVerified": true},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to check email sender verification (%s)", h.reqId, err.Error())
		return false
	}

	return count > 0
}

func (h *helper) syncSenderVerification(senders *[]email.Sender) {
	for i, sender := range *senders {
		if h.isSenderVerified(&sender) {
			(*senders)[i].AccessVerified = true
		}
	}
}

func (h *helper) getVerifiedSender() (*email.Sender, error) {
	senders, err := cubecos.GetEmailSenders()
	if err != nil {
		return nil, err
	}
	if len(senders) == 0 {
		return nil, errors.New("no email sender found")
	}

	sender := senders[0]
	if !h.isSenderVerified(&sender) {
		return nil, errors.New("email sender not verified")
	}

	return &sender, nil
}
