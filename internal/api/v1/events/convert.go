package events

import (
	"fmt"
	"net/url"

	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
)

func (h *helper) setQueryUrlToEachEvent(events *[]event.Stat) {
	for i, event := range *events {
		(*events)[i].Query = h.genEventQueryUrl(event)
	}
}

func (h *helper) genEventQueryUrl(event event.Stat) string {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = h.c.Request.Host
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/events", v1.DataCenterName)
	u.RawQuery = h.genEventQuery(event)
	return u.String()
}

func (h *helper) genEventQuery(event event.Stat) string {
	if query.IsPastRequired(h.c) {
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
		h.Period.Start,
		h.Period.Stop,
	)
}

func convertSystemSeverities(severities []string) []string {
	converted := []string{}
	for _, s := range severities {
		converted = append(
			converted,
			event.GetSeverityFullName(s),
		)
	}

	return converted
}
