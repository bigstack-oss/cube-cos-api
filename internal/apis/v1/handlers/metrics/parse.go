package metrics

import (
	"fmt"
	"strconv"
	"time"

	query "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	duration "github.com/xhit/go-str2duration"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "getDataCenterSummary":
		return h.parseWatch()
	default:
		return h.parseParams()
	}
}

func (h *helper) parseParams() error {
	parsers := []func() error{
		h.parseView, h.parseMetric, h.parseEntityType, h.parseEntityId,
		h.parsePast, h.parseAggregateWindow, h.parsePeriod,
		h.parseRank, h.parseLimit, h.parseWatch,
	}

	for _, parse := range parsers {
		err := parse()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *helper) parseView() error {
	h.viewType = h.c.Param("viewType")
	if !cubecos.IsMetricReportTypeValid(h.viewType) {
		return errors.ErrViewTypeInvalid
	}

	return nil
}

func (h *helper) parseMetric() error {
	h.metricType = h.c.Param("metricType")
	if !cubecos.IsValidMetricType(h.metricType) {
		return errors.ErrMetricTypeInvalid
	}

	return nil
}

func (h *helper) parseEntityType() error {
	h.entityType = h.c.Param("entityType")
	if !cubecos.IsEntityTypeValid(h.entityType) {
		return errors.ErrEntityTypeInvalid
	}

	return nil
}

func (h *helper) parseEntityId() error {
	h.entityId = h.c.Param("entityId")
	if h.entityId == "" {
		return nil
	}

	switch h.entityType {
	case "hosts":
		h.entityType = "host"
	case "vms":
		h.entityType = "vm"
	}

	return nil
}

func (h *helper) parseLimit() error {
	if h.viewType != "rank" {
		return nil
	}

	var err error
	h.limit, err = query.GetLimit(h.c)
	return err
}

func (h *helper) parsePast() error {
	var err error
	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}
	if h.past == "" {
		h.past = "1h"
	}

	_, err = duration.Str2Duration(h.past)
	if err != nil {
		return fmt.Errorf("invalid 'past' duration: %s", h.past)
	}

	return nil
}

func (h *helper) parseAggregateWindow() error {
	h.aggregateWindow = h.c.DefaultQuery("aggregateWindow", "")
	if h.aggregateWindow == "" {
		h.aggregateWindow = "1m"
	}
	_, err := duration.Str2Duration(h.aggregateWindow)
	if err != nil {
		return fmt.Errorf("invalid 'aggregateWindow' duration: %s", h.aggregateWindow)
	}

	past, err := duration.Str2Duration(h.past)
	if err != nil {
		return fmt.Errorf("invalid 'past' duration: %s", h.past)
	}
	if past > 12*time.Hour {
		h.aggregateWindow = "30m"
	}
	if past > 24*time.Hour {
		h.aggregateWindow = "1h"
	}

	return nil
}

func (h *helper) parsePeriod() error {
	if h.viewType != "history" {
		return nil
	}

	var err error
	h.Period, err = query.GetPeriod(h.c)
	return err
}

func (h *helper) parseRank() error {
	if h.viewType != "rank" {
		return nil
	}

	var err error
	head := h.c.DefaultQuery("head", "10")
	h.head, err = strconv.Atoi(head)
	if err != nil || h.head <= 0 {
		return fmt.Errorf("'head' should be an integer which greater than 0: %s", head)
	}

	tail := h.c.DefaultQuery("tail", "10")
	h.tail, err = strconv.Atoi(tail)
	if err != nil || h.tail <= 0 {
		return fmt.Errorf("'tail' should be an integer which greater than 0: %s", tail)
	}

	return nil
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	return err
}

func (h *helper) isPastRequired() bool {
	_, found := h.c.GetQuery("past")
	return found
}

func (h *helper) genTimeDuration() string {
	if h.isPastRequired() {
		return fmt.Sprintf("start: -%s", h.past)
	}

	return fmt.Sprintf(
		"start: %s, stop: %s",
		h.Period.Start,
		h.Period.Stop,
	)
}
