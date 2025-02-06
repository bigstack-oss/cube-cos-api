package healths

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c *gin.Context
	period
	definition.Page
}

type period struct {
	start string
	stop  string
}

type data struct {
	Health          cubecos.HealthCheckResult `json:"health"`
	definition.Page `json:"page"`
}

func (p period) StartAsTime() time.Time {
	t, err := time.Parse(time.RFC3339, p.start)
	if err != nil {
		return time.Now()
	}

	return t
}

func (p period) StopAsTime() time.Time {
	t, err := time.Parse(time.RFC3339, p.stop)
	if err != nil {
		return time.Now()
	}

	return t
}

func initReqHelper(c *gin.Context) (*helper, error) {
	h := &helper{c: c}

	err := h.parsePeriod()
	if err != nil {
		return nil, err
	}

	err = h.parsePage()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) parsePeriod() error {
	qStart := h.c.DefaultQuery("start", definition.TimeRFC3339(-24*time.Hour))
	start, err := time.Parse(time.RFC3339, qStart)
	if err != nil {
		return fmt.Errorf("'start' time format should be aligned with RFC3339: %s", qStart)
	}

	qStop := h.c.DefaultQuery("stop", definition.TimeNowRFC3339())
	stop, err := time.Parse(time.RFC3339, qStop)
	if err != nil {
		return fmt.Errorf("'stop' time format should be aligned with RFC3339: %s", qStop)
	}

	if stop.Before(start) {
		return fmt.Errorf("'stop' time should be after 'start' time(start: %s, stop: %s)", start, stop)
	}

	h.period = period{
		start: definition.TimeUTC(start),
		stop:  definition.TimeUTC(stop),
	}
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

	var err error
	h.Page.Number, err = strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", num)
	}

	h.Page.Size, err = strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", size)
	}

	return nil
}
