package triggers

type ApplyOptions struct {
	Name            string `json:"name" bson:"name"`
	Description     string `json:"description" bson:"description"`
	ApplyAttributes `json:"attributes" bson:"attributes"`
	ApplyResponse   `json:"response" bson:"response"`
}

type ApplyAttributes struct {
	AlertTypes []string `json:"alertTypes" bson:"alertTypes"`
	EventIds   []string `json:"eventIds" bson:"eventIds"`
	Severities []string `json:"severities" bson:"severities"`
	Categories []string `json:"categories" bson:"categories"`
}

type ApplyResponse struct {
	Script `json:"script" bson:"script"`
	Emails []string `json:"emails" bson:"emails"`
	Slacks []string `json:"slacks" bson:"slacks"`
}

type Script struct {
	FilePath string `json:"filePath" bson:"filePath"`
	Content  string `json:"content" bson:"content"`
}
