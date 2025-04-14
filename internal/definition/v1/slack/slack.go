package slack

import "github.com/bigstack-oss/cube-cos-api/internal/status"

const (
	ChannelCollection = "slackChannels"
)

type Options struct {
	Channels []Channel `json:"channels" bson:"channels"`
}

type Channel struct {
	Name        string          `json:"name" bson:"name"`
	URL         string          `json:"url" bson:"url"`
	Description string          `json:"description" bson:"description"`
	Enabled     bool            `json:"enabled,omitempty" bson:"enabled"`
	Status      status.Settings `json:"status" bson:"status"`
}

func (o *Options) InitOkStatus() {
	for i := range o.Channels {
		o.Channels[i].Status.InitOkStatus()
	}
}

func (c *Channel) InitUpdateStatus() {
	c.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}
