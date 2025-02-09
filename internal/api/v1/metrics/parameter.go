package metrics

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/trick"
)

func (h *helper) parseParams() error {
	parsers := []func() error{
		h.parseView, h.parseMetric, h.parseEntity,
		h.parsePeriod, h.parseRank, h.parseLimit, h.parseWatch,
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
		return errors.New("viewType should be summary, history, or rank")
	}

	return nil
}

func (h *helper) parseMetric() error {
	h.metricType = h.c.Param("metricType")
	if !cubecos.IsMetricTypeValid(h.metricType) {
		return errors.New(
			"metricType should be cpuUsage, memoryUsage, diskUsage, diskBandwidth, diskIops, diskLatency, diskReadIops, diskWriteIops, networkTrafficIn, or networkTrafficOut",
		)
	}

	return nil
}

func (h *helper) parseEntity() error {
	h.entityType = h.c.Param("entityType")
	if !cubecos.IsEntityTypeValid(h.entityType) {
		return errors.New("entityType should be hosts or vms")
	}

	return nil
}

func (h *helper) parseLimit() error {
	if h.viewType != "rank" {
		return nil
	}

	var err error
	h.limit, err = strconv.Atoi(h.c.DefaultQuery("limit", "10"))
	if err != nil || h.limit <= 0 {
		return errors.New("limit should be an integer and greater than 0")
	}

	return nil
}

func (h *helper) parsePeriod() error {
	if h.viewType != "history" {
		return nil
	}

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

	h.Period = definition.Period{
		Start: definition.TimeUTC(trick.Minus2MinsOnMetricStart(start)),
		Stop:  definition.TimeUTC(stop),
	}
	return nil
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
	h.watch, err = parseWatch(h.c)
	if err != nil {
		return errors.New("watch parameter is invalid, it should be true or false if provided")
	}

	return nil
}
