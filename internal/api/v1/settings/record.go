package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) eraseSenderPassword(senders *[]email.Sender) {
	for i := range *senders {
		(*senders)[i].ErasePassword()
	}
}

func (h *helper) updateSetting() {
	h.addReqRecord(*h.task)
	reqQueue.Add(h.task)
}

func getSlackChannel(name string) (*slack.Channel, error) {
	h := mongo.GetGlobalHelper()
	resp, err := h.Get(
		v1.SettingsDB(),
		slack.ChannelCollection,
		bson.M{"name": name},
	)
	if err != nil {
		log.Errorf("settings: failed to get slack channel (%s)", err.Error())
		return nil, err
	}

	channel := slack.Channel{}
	err = resp.Decode(&channel)
	if err != nil {
		log.Errorf("settings: failed to decode slack channel (%s)", err.Error())
		return nil, err
	}

	return &channel, nil
}

func (h *helper) isSenderExist() bool {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	return policy.HasSender(h.c.Param("senderHost"))
}

func isRecipientExist(recipient string) bool {
	h := mongo.GetGlobalHelper()
	count, err := h.GetCount(
		v1.SettingsDB(),
		email.RecipientCollection,
		bson.M{"address": recipient},
	)
	if err != nil {
		log.Errorf("settings: failed to get count of email recipient (%s)", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isTitlePrefixUpdating() bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "titlePrefix"},
	)
	if err != nil {
		log.Errorf("settings: failed to check title prefix update: %v", err)
		return false
	}

	return count > 0
}

func (h *helper) isEmailRecipientUpdating(recipient *email.Recipient) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{
			"type":    "emailRecipient",
			"address": recipient.Address,
		},
	)
	if err != nil {
		log.Errorf("settings: failed to get email count: %s", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isEmailSenderUpdating(sender *email.Sender) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{
			"type": "emailSender",
			"key":  sender.Host,
		},
	)
	if err != nil {
		log.Errorf("settings: failed to get email count: %s", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isSlackUpdating(channel *slack.Channel) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{
			"type": "slackChannel",
			"url":  channel.URL,
		},
	)
	if err != nil {
		log.Errorf("settings: failed to get slack count: %s", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) updateSettingTask() error {
	return h.mongo.DeleteOne(
		setting.DB,
		setting.ReqCollection,
		h.genTaskFilter(),
	)
}

func (h *helper) genTaskFilter() bson.M {
	return bson.M{
		"type": h.task.Type,
		"key":  h.task.Key,
	}
}

func (h *helper) resetAccessVerification() error {
	return h.mongo.UpdateMany(
		setting.DB,
		email.SenderCollection,
		bson.M{"host": h.c.Param("senderHost")},
		bson.M{"$set": bson.M{"accessVerified": false}},
	)
}
