package slack

const (
	ChannelCollection = "slackChannels"
)

type Options struct {
	Channels []Channel `json:"channels" bson:"channels"`
}

type Channel struct {
	Name        string `json:"name" bson:"name"`
	URL         string `json:"url" bson:"url"`
	Description string `json:"description" bson:"description"`
	Enabled     bool   `json:"enabled" bson:"enabled"`
}
