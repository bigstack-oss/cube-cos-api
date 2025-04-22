package triggers

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
)

// M1 TODO:
// the function below will took a bit longer to run due to the hex_sdk
// need to incrase the perf of the function by async data preloading
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
	recipients, err := cubecos.GetEmailRecipients()
	if err != nil {
		return
	}

	trigger.Emails = recipients
}

func setSlackChannelsToTrigger(trigger *trigger.Options) {
	channels, err := cubecos.GetSlackChannels()
	if err != nil {
		return
	}

	trigger.Slacks = convertToApiChannels(channels)
}

func convertToApiChannels(channels []slack.CosChannel) []slack.ApiChannel {
	apiChannels := []slack.ApiChannel{}

	for _, channel := range channels {
		apiChannels = append(
			apiChannels,
			slack.ApiChannel{
				Name:        channel.Channel,
				URL:         channel.URL,
				Description: channel.Description,
			},
		)
	}

	return apiChannels
}
