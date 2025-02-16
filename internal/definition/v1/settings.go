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

func (es EmailSender) Collection() string {
	return "emailSender"
}

func SettingsDB() string {
	return Settings
}
