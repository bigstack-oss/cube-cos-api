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
	definition.Page
}

type period struct {
	start string
	stop  string
}

func initReqHelper(c *gin.Context) (*helper, error) {
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
	num := h.c.DefaultQuery("pageNum", "")
	size := h.c.DefaultQuery("pageSize", "")
	if !isPageRequired(num, size) {
		return nil
	}

	if num == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if size == "" {
		return fmt.Errorf("pageSize should greater than 0 if pageNum is provided")
	}

	intPageNum, err := strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", num)
	}

	intPageSize, err := strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", size)
	}

	h.Page = definition.Page{Number: intPageNum, Size: intPageSize}
	return nil
}
