package events

import (
	"math"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func isPageRequired(page, pageSize string) bool {
	return page != "" || pageSize != ""
}

func (h *helper) isPageRequired() bool {
	return h.Page.Number > 0 || h.Page.Size > 0
}

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

func (h *helper) genPageInfo(events []v1.Event) (v1.Page, error) {
	if !h.isPageRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(events),
			TotalItemCount: int64(len(events)),
		}, nil
	}

	totalCounts, totalPages, err := h.getAmountDetails()
	if err != nil {
		return v1.Page{}, err
	}

	return v1.Page{
		Total:          totalPages,
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: totalCounts,
	}, nil
}

func (h *helper) getAmountDetails() (int64, int64, error) {
	count, err := cubecos.CountEvents(h.genCountQueryStmt())
	if err != nil {
		log.Errorf("request(%s): failed to count events: %v", api.GetReqId(h.c), err)
		return 0, 0, err
	}

	return int64(count),
		int64(math.Ceil(float64(count) / float64(h.Page.Size))),
		nil
}
