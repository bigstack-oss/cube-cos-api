package triggers

type materials struct {
	Attribute `json:"attribute"`
	Response  `json:"response"`
}

type Attribute struct {
	AlertTypes []string `json:"alertTypes"`
	Severities []string `json:"severities"`
	Categories []string `json:"categories"`
	EventIds   []string `json:"eventIds"`
}

type Response struct {
	ScriptTypes   []string `json:"scriptTypes"`
	Notifications `json:"notifications"`
}

type Notifications struct {
	Emails []Email `json:"emails"`
	Slacks []Slack `json:"slacks"`
}

type Email struct {
	Address string `json:"address"`
	Note    string `json:"note"`
}

type Slack struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
}
