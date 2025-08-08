package notifications

import (
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/google/uuid"
)

const (
	Module          = "notifications"
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
	SearchId       string            `json:"-" bson:"-"`
}

type ListOpts struct {
	Limit     int64  `json:"limit"`
	Desending bool   `json:"descending"`
	Start     string `json:"start"`
	Stop      string `json:"stop"`
}

func (n *Notification) SetSearchId() {
	n.SearchId = uuid.New().String()[:8]
}

func (n *Notification) GenSearchableObject() Notification {
	return Notification{
		Id:             search.NormalizeKeyword(n.Id),
		NodeName:       search.NormalizeKeyword(n.NodeName),
		Time:           search.NormalizeKeyword(n.Time),
		AdditionalInfo: search.NormalizeMap(n.AdditionalInfo),
	}
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
