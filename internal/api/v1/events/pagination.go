package events

import (
	"math"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func isPaginationRequired(page, pageSize string) bool {
	return page != "" || pageSize != ""
}

func (h *helper) isPaginationRequired() bool {
	return h.Page.Number > 0 || h.Page.Size > 0
}

func (h *helper) genPageInfo(events []definition.Event) (definition.Page, error) {
	if !h.isPaginationRequired() {
		return definition.Page{
			Total:  1,
			Number: 1,
			Size:   len(events),
		}, nil
	}

	totalPages, err := h.countTotalPages()
	if err != nil {
		return definition.Page{}, err
	}

	return definition.Page{
		Total:  totalPages,
		Number: h.Page.Number,
		Size:   h.Page.Size,
	}, nil
}

func (h *helper) countTotalPages() (int64, error) {
	count, err := cubecos.CountEvents(h.genCountQueryStmt())
	if err != nil {
		log.Errorf("request(%s): failed to count events: %v", api.GetReqId(h.c), err)
		api.SetInternalServerError(h.c, err)
		return 0, err
	}

	return int64(math.Ceil(float64(count) / float64(h.Page.Size))), nil
}
