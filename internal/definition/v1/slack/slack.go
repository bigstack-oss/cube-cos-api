package slack

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	ChannelCollection = "slackChannels"
)

type Options struct {
	Channels []ApiChannel `json:"channels" bson:"channels"`
}

type ApiChannel struct {
	Name        string          `json:"name" bson:"name"`
	URL         string          `json:"url" bson:"url"`
	Description string          `json:"description" bson:"description"`
	Enabled     bool            `json:"enabled,omitempty" bson:"enabled"`
	Status      status.Settings `json:"status" bson:"status"`
}

type CosChannel struct {
	Channel     string `json:"channel" bson:"channel"`
	URL         string `json:"url" bson:"url"`
	Username    string `json:"username" bson:"username"`
	Workspace   string `json:"workspace" bson:"workspace"`
	Description string `json:"description" bson:"description"`
}

func (o *Options) InitOkStatus() {
	for i := range o.Channels {
		o.Channels[i].Status.InitOkStatus()
	}
}

func (c *ApiChannel) SetUpdating() {
	c.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}

func (c *ApiChannel) ToCosSchema() CosChannel {
	return CosChannel{
		Channel:     c.Name,
		URL:         c.URL,
		Username:    "",
		Workspace:   "",
		Description: c.Description,
	}
}
