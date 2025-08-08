package notifications

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type notificationPage struct {
	Notifications []notifications.Notification `json:"notifications"`
	pages.Page    `json:"page"`
}

func (h *helper) paginateNotifications(notifications []notifications.Notification) []notifications.Notification {
	if !h.page.IsRequired() {
		return notifications
	}

	left := min((h.page.Number-1)*h.page.Size, len(notifications))
	right := min(left+h.page.Size, len(notifications))
	return notifications[left:right]
}

func (h *helper) sortNotifications(notifications *[]notifications.Notification) {
	sort.Slice(*notifications, func(i, j int) bool {
		return (*notifications)[i].Time > (*notifications)[j].Time
	})
}

func (h *helper) genPageInfo(notifications []notifications.Notification) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(notifications),
			TotalItemCount: int64(len(notifications)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(notifications)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(notifications)),
	}
}
