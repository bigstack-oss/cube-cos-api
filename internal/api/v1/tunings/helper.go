package tunings

import (
	"fmt"
	"math"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	allNodes bool
	definition.Page

	watch bool
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	err := h.parsePage()
	if err != nil {
		return nil, err
	}

	h.parseScope()
	h.parseWatch()

	return h, nil
}

func (h *helper) parseScope() {
	h.allNodes = h.c.DefaultQuery("allNodes", "false") == "true"
}

func (h *helper) parsePage() error {
	num := h.c.DefaultQuery("pageNum", "")
	size := h.c.DefaultQuery("pageSize", "")
	if !isPageReceived(num, size) {
		return nil
	}

	if num == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if size == "" {
		return fmt.Errorf("pageSize should be provided if pageNum is provided")
	}

	var err error
	h.Page.Number, err = strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", num)
	}

	h.Page.Size, err = strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", size)
	}

	if h.Page.Number <= 0 {
		return fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if h.Page.Size <= 0 {
		return fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return nil
}

func (h *helper) parseWatch() {
	h.watch = h.c.DefaultQuery("watch", "false") == "true"
}

func (h *helper) ListTunings() (*data, error) {
	tunings, err := cubecos.ListTunings(definition.ListTuningOptions{AllNodes: h.allNodes})
	if err != nil {
		log.Errorf("request(%s): failed to get tunings: %s", api.GetReqId(h.c), err.Error())
		api.SetInternalServerError(h.c, err)
		return nil, err
	}

	pagedTunings, err := h.paginateTunings(tunings)
	if err != nil {
		log.Errorf("request(%s): failed to paginate tunings: %s", api.GetReqId(h.c), err.Error())
		api.SetInternalServerError(h.c, err)
		return nil, err
	}

	page, err := h.genPageInfo(tunings)
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %s", api.GetReqId(h.c), err.Error())
		api.SetInternalServerError(h.c, err)
		return nil, err
	}

	return &data{
		Tunings: pagedTunings,
		Page:    page,
	}, nil
}

func (h *helper) paginateTunings(tunings []definition.Tuning) ([]definition.Tuning, error) {
	if !h.Page.IsRequired() {
		return tunings, nil
	}

	left := (h.Page.Number - 1) * h.Page.Size
	if left > len(tunings) {
		left = len(tunings)
	}

	right := left + h.Page.Size
	if right > len(tunings) {
		right = len(tunings)
	}

	return tunings[left:right], nil
}

func (h *helper) genPageInfo(tunings []definition.Tuning) (definition.Page, error) {
	if !h.Page.IsRequired() {
		return definition.Page{
			Total:  1,
			Number: 1,
			Size:   len(tunings),
		}, nil
	}

	return definition.Page{
		Total:  int64(math.Ceil(float64(len(tunings)) / float64(h.Page.Size))),
		Number: h.Page.Number,
		Size:   h.Page.Size,
	}, nil
}

func isPageReceived(num, size string) bool {
	return num != "" || size != ""
}
