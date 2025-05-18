package settings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

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
