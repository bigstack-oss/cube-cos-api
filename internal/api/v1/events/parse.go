package events

import (
	"errors"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
)

func (h *helper) parseEventListingParams() error {
	err := h.parseType()
	if err != nil {
		return err
	}

	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}

	h.Period, err = query.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.Page, err = query.GetPage(h.c)
	if err != nil {
		return err
	}

	err = h.parseFilterConditions()
	if err != nil {
		return err
	}

	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseEventAbstractParams() error {
	err := h.parseType()
	if err != nil {
		return err
	}

	h.limit, err = query.GetLimit(h.c)
	if err != nil {
		return err
	}

	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseEventRankParams() error {
	err := h.parseType()
	if err != nil {
		return err
	}

	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}

	h.Period, err = query.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.limit, err = query.GetLimit(h.c)
	if err != nil {
		return err
	}

	err = h.parseFilterConditions()
	if err != nil {
		return err
	}

	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseEventFilterConditions() error {
	var err error
	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}

	h.Period, err = query.GetPeriod(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseType() error {
	query := h.c.DefaultQuery("type", "")
	if !events.IsValidType(query) {
		return errors.New(
			"'type' can't be null and should be one of 'system', 'host', or 'instance'",
		)
	}

	h.eventType = query
	return nil
}

func (h *helper) parseFilterConditions() error {
	queries := h.c.Request.URL.Query()
	for _, condition := range v1.GetFilterConditions() {
		value, found := queries[condition]
		if !found {
			continue
		}
		if len(value) == 0 {
			continue
		}
		if value[0] == "" {
			continue
		}

		switch condition {
		case "id":
			h.eventId = value[0]
		case "category":
			h.category = strings.ToUpper(value[0])
		case "severity":
			h.severity = v1.SeverityShortName(value[0])
		case "host":
			h.host = value[0]
		case "instance":
			h.instance = value[0]
		case "keyword":
			h.keyword = value[0]
		}
	}

	return nil
}
