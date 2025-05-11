package settings

import (
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
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

func initUpdateStatus() status.Settings {
	return status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
	}
}
