package triggers

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
)

func (h *helper) syncSelectableResponseItems(trigger *trigger.Options) {
	trigger.InitResponse()

	setEmailRecipientsToTrigger(trigger)
	if trigger.HasEmailRecipients() {
		trigger.Response.Types = append(
			trigger.Response.Types,
			"email",
		)
	}

	setSlackChannelsToTrigger(trigger)
	if trigger.HasSlackChannels() {
		trigger.Response.Types = append(
			trigger.Response.Types,
			"slack",
		)
	}
}

func setEmailRecipientsToTrigger(trigger *trigger.Options) {
	recipients, err := v1.GetEmailRecipients()
	if err != nil {
		return
	}

	trigger.Emails = recipients
}

func setSlackChannelsToTrigger(trigger *trigger.Options) {
	channels, err := v1.GetSlackChannels()
	if err != nil {
		return
	}

	trigger.Slacks = channels
}
