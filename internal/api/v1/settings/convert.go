package settings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

func convertToApiSlackChannels(channels []slack.CosChannel) []slack.ApiChannel {
	apiChannels := []slack.ApiChannel{}

	for i, channel := range channels {
		apiChannels[i] = slack.ApiChannel{
			Name:        channel.Channel,
			URL:         channel.URL,
			Description: channel.Description,
		}
	}

	return apiChannels
}
