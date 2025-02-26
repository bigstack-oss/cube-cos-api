package healths

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	duration "github.com/xhit/go-str2duration"
)

type helper struct {
	c *gin.Context

	service string
	module  string
	handler string

	period
	past string

	definition.Page
	watch bool
}

type period struct {
	start string
	stop  string
}

func (p period) StartTime() time.Time {
	t, err := time.Parse(time.RFC3339, p.start)
	if err != nil {
		return time.Now().Add(-time.Hour)
	}

	return t
}

func (p period) StopTime() time.Time {
	t, err := time.Parse(time.RFC3339, p.stop)
	if err != nil {
		return time.Now()
	}

	return t
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	switch h.handler {
	case "getHealthSummary":
		return h.parseSummaryParams()
	case "getHealthHistoryOfService":
		return h.parseServiceHealthParams()
	case "getHealthHistoryOfModule":
		return h.parseModuleHealthParams()
	}

	return h, nil
}

func (h *helper) parseSummaryParams() (*helper, error) {
	h.parseWatch()
	return h, nil
}

func (h *helper) parseServiceHealthParams() (*helper, error) {
	h.parseWatch()

	err := h.parsePast()
	if err != nil {
		return nil, err
	}

	err = h.parsePeriod()
	if err != nil {
		return nil, err
	}

	h.service = h.c.Param("serviceType")
	if !cubecos.IsValidService(h.service) {
		return nil, fmt.Errorf("invalid serviceType: %s", h.service)
	}

	return h, nil
}

func (h *helper) parseModuleHealthParams() (*helper, error) {
	h.parseWatch()

	err := h.parsePast()
	if err != nil {
		return nil, err
	}

	err = h.parsePeriod()
	if err != nil {
		return nil, err
	}

	h.service = h.c.Param("serviceType")
	if !cubecos.IsValidService(h.service) {
		return nil, fmt.Errorf("invalid serviceType: %s", h.service)
	}

	h.module = h.c.Param("moduleType")
	if !cubecos.IsValidServiceAndModule(h.service, h.module) {
		return nil, fmt.Errorf("invalid serviceType' %s' or module '%s'", h.service, h.module)
	}

	return h, nil
}

func (h *helper) parsePeriod() error {
	if h.arePeriodAndPastRequired() {
		return fmt.Errorf("'past' and 'start'/'stop' cannot be used together")
	}

	qStart := h.c.DefaultQuery("start", definition.TimeRFC3339(-time.Hour))
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

func (h *helper) parseWatch() {
	h.watch = h.c.DefaultQuery("watch", "false") == "true"
}

func (h *helper) parsePast() error {
	h.past = h.c.DefaultQuery("past", "")
	if h.past == "" {
		return nil
	}

	_, err := duration.Str2Duration(h.past)
	if err != nil {
		return fmt.Errorf("invalid 'past' duration: %s", h.past)
	}

	return nil
}

func (h *helper) arePeriodAndPastRequired() bool {
	return h.isPeriodRequired() && h.isPastRequired()
}

func (h *helper) isPeriodRequired() bool {
	return h.c.DefaultQuery("stop", "") != "" || h.c.DefaultQuery("start", "") != ""
}

func (h *helper) isPastRequired() bool {
	return h.c.DefaultQuery("past", "") != ""
}
