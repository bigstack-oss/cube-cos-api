package notifications

import (
	"sync"
	"time"
)

const (
	Db              = "notifications"
	ToastCollection = "toasts"
)

var (
	cache = sync.Map{}
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

func GetCacheById(id string) (Notification, bool) {
	notification, ok := cache.Load(id)
	if ok {
		return notification.(Notification), true
	}

	return Notification{}, false
}

func SetCacheById(id string, notification Notification) {
	cache.Store(id, notification)
}

func DeleteCacheById(id string) {
	cache.Delete(id)
}
