package events

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c         *gin.Context
	eventType string
	period
	page
}

type period struct {
	start string
	stop  string
}

type page struct {
	Total  int64 `json:"total"`
	Number int   `json:"number"`
	Size   int   `json:"size"`
}

func initHelperByQueryParams(c *gin.Context) (*helper, error) {
	h := &helper{c: c}

	err := h.parseType()
	if err != nil {
		return nil, err
	}

	err = h.parsePeriod()
	if err != nil {
		return nil, err
	}

	err = h.parsePage()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) parseType() error {
	t := h.c.DefaultQuery("type", "")
	if !cubecos.IsEventTypeValid(t) {
		return fmt.Errorf(
			"'type' can't be null and should be one of 'system', 'host', or 'instance': %s",
			t,
		)
	}

	h.eventType = t
	return nil
}

func (h *helper) parsePeriod() error {
	start := h.c.DefaultQuery("start", definition.TimeRFC3339(-24*time.Hour))
	_, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return fmt.Errorf("'start' time format should be aligned with RFC3339: %s", start)
	}

	stop := h.c.DefaultQuery("stop", definition.TimeNowRFC3339())
	_, err = time.Parse(time.RFC3339, stop)
	if err != nil {
		return fmt.Errorf("'stop' time format should be aligned with RFC3339: %s", stop)
	}

	h.period = period{start: start, stop: stop}
	return nil
}

func (h *helper) parsePage() error {
	pageNum := h.c.DefaultQuery("pageNum", "")
	pageSize := h.c.DefaultQuery("pageSize", "")
	if !isPaginationRequired(pageNum, pageSize) {
		return nil
	}

	if pageNum == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if pageSize == "" {
		return fmt.Errorf("pageSize should greater than 0 if pageNum is provided")
	}

	intPageNum, err := strconv.Atoi(pageNum)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", pageNum)
	}

	intPageSize, err := strconv.Atoi(pageSize)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", pageSize)
	}

	h.page = page{Number: intPageNum, Size: intPageSize}
	return nil
}
