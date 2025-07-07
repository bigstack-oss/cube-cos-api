package events

import (
	"errors"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listEvents":
		return h.parseEventListingParams()
	case "listPredefinedEvents":
		return h.parsePredefinedEventParams()
	case "listEventAbstract":
		return h.parseEventAbstractParams()
	case "getEventRank":
		return h.parseEventRankParams()
	case "getEventFilterConditions":
		return h.parseEventFilterConditions()
	}

	return nil
}

func (h *helper) parseEventListingParams() error {
	err := h.parseType()
	if err != nil {
		return err
	}

	h.past, err = queries.GetPast(h.c)
	if err != nil {
		return err
	}

	h.period, err = queries.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.page, err = queries.GetPage(h.c)
	if err != nil {
		return err
	}

	err = h.parseFilterConditions()
	if err != nil {
		return err
	}

	h.watch, err = queries.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parsePredefinedEventParams() error {
	return h.parsePredefinedFilterConditions()
}

func (h *helper) parseEventAbstractParams() error {
	err := h.parseType()
	if err != nil {
		return err
	}

	h.limit, err = queries.GetLimit(h.c)
	if err != nil {
		return err
	}

	h.watch, err = queries.GetWatch(h.c)
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

	h.past, err = queries.GetPast(h.c)
	if err != nil {
		return err
	}

	h.period, err = queries.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.limit, err = queries.GetLimit(h.c)
	if err != nil {
		return err
	}

	err = h.parseFilterConditions()
	if err != nil {
		return err
	}

	h.watch, err = queries.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseEventFilterConditions() error {
	var err error
	h.past, err = queries.GetPast(h.c)
	if err != nil {
		return err
	}

	h.period, err = queries.GetPeriod(h.c)
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
	for _, condition := range events.GetFilterConditions() {
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
		case "categories":
			h.categories = h.c.QueryArray("categories")
		case "severity":
			h.severity = events.GetSeverityShortName(value[0])
		case "severities":
			h.severities = events.GetSeverityFullNames(h.c.QueryArray("severities"))
		case "host":
			h.host = value[0]
		case "hosts":
			h.hosts = h.c.QueryArray("hosts")
		case "instance":
			h.instance = value[0]
		case "instances":
			h.instances = h.c.QueryArray("instances")
		case "keyword":
			h.keyword = value[0]
		}
	}

	return nil
}

func (h *helper) parsePredefinedFilterConditions() error {
	queries := h.c.Request.URL.Query()
	for _, condition := range events.GetFilterConditions() {
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
		case "types":
			h.eventTypes = h.c.QueryArray("types")
		case "categories":
			h.categories = h.c.QueryArray("categories")
		case "severities":
			h.severities = h.c.QueryArray("severities")
		}
	}

	return nil
}
