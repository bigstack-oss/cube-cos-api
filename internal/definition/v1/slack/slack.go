package slack

const (
	slackChannels = "slackChannels"
)

type Options struct {
	Channels []Channel `json:"channels" bson:"channels"`
}

type Channel struct {
	Name    string `json:"name" bson:"name"`
	URL     string `json:"url" bson:"url"`
	Deleted bool   `json:"-" bson:"deleted"`
}

func ChannelCollection() string {
	return slackChannels
}
