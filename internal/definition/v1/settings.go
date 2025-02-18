package v1

const (
	Settings        = "settings"
	EmailSenders    = "emailSenders"
	EmailRecipients = "emailRecipients"
	SlackWebhooks   = "slackWebhooks"
)

type Setting struct {
	EmailSenders    []EmailSender    `json:"emailSenders" bson:"emailSenders"`
	EmailRecipients []EmailRecipient `json:"emailRecipients" bson:"emailRecipients"`
	SlackWebhooks   []SlackWebhook   `json:"slackWebhooks" bson:"slackWebhooks"`
}

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
	return EmailSenders
}

func EmailRecipientCollection() string {
	return EmailRecipients
}

func SlackWebhookCollection() string {
	return SlackWebhooks
}

func SettingsDB() string {
	return Settings
}
