package events

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	duration "github.com/xhit/go-str2duration"
	log "go-micro.dev/v5/logger"
)

var (
	filterConditions = []string{
		"id",
		"category",
		"severity",
		"host",
		"instance",
		"keyword",
	}
)

type helper struct {
	c       *gin.Context
	handler string

	eventType string
	eventId   string
	category  string
	severity  string
	host      string
	instance  string
	keyword   string

	period
	past string

	definition.Page
	limit int

	watch bool
}

type period struct {
	start string
	stop  string
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}

	switch handler {
	case "getEvents":
		return h.parseEventListingParams()
	case "getEventAbstract":
		return h.parseEventAbstractParams()
	case "genEventRank":
		return h.parseEventRankParams()
	case "getEventFilterConditions":
		return h.parseEventFilterConditions()
	}

	return nil, errors.New("no internal function supported")
}

func (h *helper) parseEventListingParams() (*helper, error) {
	err := h.parseType()
	if err != nil {
		return nil, err
	}

	err = h.parsePast()
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

	err = h.parseFilterConditions()
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

func (h *helper) parseEventRankParams() (*helper, error) {
	err := h.parseType()
	if err != nil {
		return nil, err
	}

	err = h.parsePast()
	if err != nil {
		return nil, err
	}

	err = h.parsePeriod()
	if err != nil {
		return nil, err
	}

	err = h.parseLimit()
	if err != nil {
		return nil, err
	}

	err = h.parseFilterConditions()
	if err != nil {
		return nil, err
	}

	err = h.parseWatch()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) parseEventFilterConditions() (*helper, error) {
	err := h.parsePast()
	if err != nil {
		return nil, err
	}

	err = h.parsePeriod()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) parseType() error {
	t := h.c.DefaultQuery("type", "")
	if !cubecos.IsEventTypeValid(t) {
		return errors.New(
			"'type' can't be null and should be one of 'system', 'host', or 'instance'",
		)
	}

	h.eventType = t
	return nil
}

func (h *helper) parsePeriod() error {
	if h.arePeriodAndPastRequired() {
		return fmt.Errorf("'past' and 'start'/'stop' cannot be used together")
	}

	qStart := h.c.DefaultQuery("start", definition.TimeRFC3339(-24*time.Hour))
	start, err := time.Parse(definition.RFC3339, qStart)
	if err != nil {
		return fmt.Errorf("'start' time format should be aligned with RFC3339: %s", qStart)
	}

	qStop := h.c.DefaultQuery("stop", definition.TimeNowRFC3339())
	stop, err := time.Parse(definition.RFC3339, qStop)
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

func (h *helper) parseFilterConditions() error {
	receivedParams := h.c.Request.URL.Query()
	for _, c := range filterConditions {
		v, found := receivedParams[c]
		if !found {
			continue
		}
		if len(v) == 0 {
			continue
		}
		if v[0] == "" {
			continue
		}

		switch c {
		case "id":
			h.eventId = v[0]
		case "category":
			h.category = strings.ToUpper(v[0])
		case "severity":
			h.severity = definition.SeverityShortName(v[0])
		case "host":
			h.host = v[0]
		case "instance":
			h.instance = v[0]
		case "keyword":
			h.keyword = v[0]
		}
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

func (h *helper) genEventRank() (*data, error) {
	stmt, err := h.genRankStmt()
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(h.c), err)
		return nil, err
	}

	rank, err := cubecos.GetEventRank(stmt)
	if err != nil {
		log.Errorf("request(%s): failed to get events: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	h.setQueryUrlToEachEvent(&rank)
	return &data{
		Events: rank,
		Limit: &definition.Limit{
			Number:      h.limit,
			Description: fmt.Sprintf("The top %d event IDs with the highest proportion", len(rank)),
		},
	}, nil
}

func (h *helper) setQueryUrlToEachEvent(events *[]definition.EventStat) {
	for i, event := range *events {
		(*events)[i].Query = h.genEventQueryUrl(event)
	}
}

func (h *helper) genEventQueryUrl(event definition.EventStat) string {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = h.c.Request.Host
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/events", definition.DataCenterName)
	u.RawQuery = h.genEventQuery(event)
	return u.String()
}

func (h *helper) genEventQuery(event definition.EventStat) string {
	if h.isPastRequired() {
		return fmt.Sprintf(
			"type=%s&id=%s&past=%s&pageNum=1&pageSize=20",
			h.eventType,
			event.Id,
			h.past,
		)
	}

	return fmt.Sprintf(
		"type=%s&id=%s&start=%s&stop=%s&pageNum=1&pageSize=20",
		h.eventType,
		event.Id,
		h.period.start,
		h.period.stop,
	)
}

func (h *helper) genEventFilterConditions() (*definition.EventFilter, error) {
	systemCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("system", "category"))
	if err != nil {
		log.Errorf("request(%s): failed to get system categories: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	systemSeverities, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("system", "severity"))
	if err != nil {
		log.Errorf("request(%s): failed to get system severities: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	hostCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("host", "category"))
	if err != nil {
		log.Errorf("request(%s): failed to get host categories: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	hostNames, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("host", "host"))
	if err != nil {
		log.Errorf("request(%s): failed to get host names: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	instanceCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("instance", "category"))
	if err != nil {
		log.Errorf("request(%s): failed to get instance categories: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	instanceIds, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("instance", "instance"))
	if err != nil {
		log.Errorf("request(%s): failed to get instances: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	return &definition.EventFilter{
		System: definition.SystemFilter{
			Categories: systemCategories,
			Severities: convertSystemSeverities(systemSeverities),
		},
		Host: definition.HostFilter{
			Categories: hostCategories,
			Names:      hostNames,
		},
		Instance: definition.InstanceFilter{
			Categories: instanceCategories,
			Ids:        instanceIds,
		},
	}, nil
}

func convertSystemSeverities(severities []string) []string {
	converted := []string{}
	for _, s := range severities {
		converted = append(
			converted,
			definition.SeverityFullName(s),
		)
	}

	return converted
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
	return h.past != ""
}
