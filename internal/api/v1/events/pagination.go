package events

import (
	"math"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
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

func (h *helper) paginateEvents(events []events.Options) ([]events.Options, error) {
	if !h.Page.IsRequired() {
		return events, nil
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(events))
	right := min(left+h.Page.Size, len(events))
	return events[left:right], nil
}

func (h *helper) genPageInfo(events []events.Options) (v1.Page, error) {
	if !h.Page.IsRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(events),
			TotalItemCount: int64(len(events)),
		}, nil
	}

	totalCounts, totalPages := h.getAmountDetails(events)
	return v1.Page{
		Total:          totalPages,
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: totalCounts,
	}, nil
}

func (h *helper) getAmountDetails(events []events.Options) (int64, int64) {
	count := len(events)
	return int64(count),
		int64(math.Ceil(float64(count) / float64(h.Page.Size)))
}
