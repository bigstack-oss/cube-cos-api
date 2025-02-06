package events

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c         *gin.Context
	handler   string
	eventType string

	period
	definition.Page
	limit int

	watch bool
}

type period struct {
	start string
	stop  string
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c}

	switch handler {
	case "getEvents":
		h.handler = "getEvents"
		return h.parseEventListingParams()
	case "getEventAbstract":
		h.handler = "getEventAbstract"
		return h.parseEventAbstractParams()
	}

	return nil, errors.New("no internal function supported")
}

func (h *helper) parseEventListingParams() (*helper, error) {
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

	err = h.parseWatch()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) parseEventAbstractParams() (*helper, error) {
	err := h.parseType()
	if err != nil {
		return nil, err
	}

	err = h.parseLimit()
	if err != nil {
		return nil, err
	}

	err = h.parseWatch()
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

func (h *helper) parseLimit() error {
	var err error
	limit := h.c.DefaultQuery("limit", "10")
	h.limit, err = strconv.Atoi(limit)
	if err != nil {
		return err
	}

	if h.limit <= 0 {
		return fmt.Errorf("limit should be greater than 0")
	}

	return nil
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = api.ParseWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) genEvents() (*data, error) {
	stmt := h.genListingStmt()
	events, err := cubecos.GetEvents(stmt)
	if err != nil {
		log.Errorf("request(%s): failed to get events: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	page, err := h.genPageInfo(events)
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	return &data{
		Events: events,
		Page:   &page,
	}, nil
}

func (h *helper) genEventAbstract() (*data, error) {
	stmt := h.genAbstractStmt()
	events, err := cubecos.GetEvents(stmt)
	if err != nil {
		log.Errorf("request(%s): failed to get events: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	return &data{
		Events: events,
		Limit: &definition.Limit{
			Number:      h.limit,
			Description: fmt.Sprintf("the top %d recent events", h.limit),
		},
	}, nil
}
