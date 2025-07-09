package events

import (
	"math"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isTypesRequired() bool {
	return len(h.eventTypes) > 0
}

func (h *helper) isEventIdsRequired() bool {
	return len(h.eventIds) > 0
}

func (h *helper) isCategoriesRequired() bool {
	return len(h.categories) > 0
}

func (h *helper) isIdRequired() bool {
	return h.eventId != ""
}

func (h *helper) isSeveritiesRequired() bool {
	return len(h.severities) > 0
}

func (h *helper) isHostsRequired() bool {
	return len(h.hosts) > 0
}

func (h *helper) isInstancesRequired() bool {
	return len(h.instances) > 0
}

func (h *helper) paginateEvents(events []events.Event) ([]events.Event, error) {
	if !h.page.IsRequired() {
		return events, nil
	}

	left := min((h.page.Number-1)*h.page.Size, len(events))
	right := min(left+h.page.Size, len(events))
	return events[left:right], nil
}

func (h *helper) genPageInfo(events []events.Event) (pages.Page, error) {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(events),
			TotalItemCount: int64(len(events)),
		}, nil
	}

	totalCounts, totalPages := h.getAmountDetails(events)
	return pages.Page{
		Total:          totalPages,
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: totalCounts,
	}, nil
}

func (h *helper) getAmountDetails(events []events.Event) (int64, int64) {
	count := len(events)
	return int64(count),
		int64(math.Ceil(float64(count) / float64(h.page.Size)))
}
