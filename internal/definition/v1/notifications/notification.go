package notifications

import "time"

const (
	Db     = "notifications"
	Toasts = "toasts"
)

type Notification struct {
	Id             string            `json:"id" bson:"id"`
	NodeName       string            `json:"nodeName" bson:"nodeName"`
	Time           string            `json:"time" bson:"time"`
	AdditionalInfo map[string]string `json:"additionalInfo" bson:"additionalInfo"`
}

type ListOpts struct {
	Limit     int64     `json:"limit"`
	Desending bool      `json:"descending"`
	Start     time.Time `json:"start"`
	Stop      time.Time `json:"stop"`
}
