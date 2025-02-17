package v1

const (
	Settings = "settings"
)

type EmailSender struct {
	Deleted bool `json:"-" bson:"deleted"`

	Host     string `json:"host" bson:"host"`
	Port     int    `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	From     string `json:"from" bson:"from"`

	Note string `json:"note,omitempty" bson:"note,omitempty"`
}

type EmailRecipient struct {
	ID      string `json:"id" bson:"id"`
	Deleted bool   `json:"-" bson:"deleted"`

	To   []string `json:"to" bson:"to"`
	Note string   `json:"note,omitempty" bson:"note,omitempty"`
}

type SlackWebhook struct {
	ID      string `json:"id" bson:"id"`
	Deleted bool   `json:"-" bson:"deleted"`

	URL     string `json:"url" bson:"url"`
	Channel string `json:"channel" bson:"channel"`
}

func EmailSenderCollection() string {
	return "emailSender"
}

func EmailRecipientCollection() string {
	return "emailRecipient"
}

func SlackWebhookCollection() string {
	return "slackWebhook"
}

func SettingsDB() string {
	return Settings
}
