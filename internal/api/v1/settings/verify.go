package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) checkRecipientUpdate() error {
	if !h.isRecipientExist() {
		return errors.New("recipient not found")
	}

	err := email.CheckFormat(h.c.Param("recipientEmail"))
	if err != nil {
		return errors.New("recipient email format is invalid")
	}

	return nil
}

func (h *helper) isRecipientExist() bool {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	return policy.HasRecipient(h.c.Param("recipientEmail"))
}

func (h *helper) isSlackChannlExist() bool {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	channel := h.task.Slack.ConvertToCosSchema()
	if policy.HasSlackChannel(channel) {
		return true
	}

	if h.hasUpdatingSlackChannel() {
		return true
	}

	return false
}

func (h *helper) hasUpdatingSlackChannel() bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
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
