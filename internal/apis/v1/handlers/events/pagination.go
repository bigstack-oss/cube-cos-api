package events

import (
	"math"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isCategoryRequired() bool {
	return h.category != ""
}

func (h *helper) isIdRequired() bool {
	return h.eventId != ""
}

func (h *helper) isSeverityRequired() bool {
	return h.severity != ""
}

func (h *helper) isHostRequired() bool {
	return h.host != ""
}

func (h *helper) isInstanceRequired() bool {
	return h.instance != ""
}

func (h *helper) paginateEvents(events []event.Options) ([]event.Options, error) {
	if !h.page.IsRequired() {
		return events, nil
	}

	left := min((h.page.Number-1)*h.page.Size, len(events))
	right := min(left+h.page.Size, len(events))
	return events[left:right], nil
}

func (h *helper) genPageInfo(events []event.Options) (pages.Page, error) {
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

func (h *helper) getAmountDetails(events []event.Options) (int64, int64) {
	count := len(events)
	return int64(count),
		int64(math.Ceil(float64(count) / float64(h.page.Size)))
}
