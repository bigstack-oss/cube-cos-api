package notifications

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterNotifications(notifications []notifications.Notification) []notifications.Notification {
	if h.isKeywordRequired() {
		notifications = h.filteredByKeyword(notifications)
	}

	return notifications
}

func (h *helper) filteredByKeyword(list []notifications.Notification) []notifications.Notification {
	h.setSerachIds(&list)
	result, err := h.searchNotifications(list)
	if err != nil {
		log.Errorf("notifications: failed to search notification(%v)", err)
		return list
	}

	notificationMap := genNotificationMap(list)
	filtered := []notifications.Notification{}
	for _, hit := range result.Hits {
		filtered = append(filtered, notificationMap[hit.ID])
	}

	return filtered
}

func (h *helper) setSerachIds(list *[]notifications.Notification) {
	for i := range *list {
		(*list)[i].SetSearchId()
	}
}

func (h *helper) searchNotifications(list []notifications.Notification) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("notifications: failed to create notification searcher(%v)", err)
		return nil, err
	}

	for _, notification := range list {
		err := searcher.Index(notification.SearchId, notification.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	key := search.NormalizeKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(key))
}

func genNotificationMap(list []notifications.Notification) map[string]notifications.Notification {
	notificationMap := map[string]notifications.Notification{}
	for _, notification := range list {
		notificationMap[notification.SearchId] = notification
	}

	return notificationMap
}
