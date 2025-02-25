package events

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
)

const (
	convertValueToField = `rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value"`
	descByTime          = `columns: ["_time"], desc: true`
)

func (h *helper) genCountQueryStmt() string {
	query := influx.Query{}
	query.Bucket("events").
		Range(h.genTimeDuration()).
		Measurement(h.eventType)

	query = h.addFilters(query)
	return query.Count("").String()
}

func (h *helper) genFilterConditionStmt(eventType, column string) string {
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genTimeDuration()).
		Measurement(eventType).
		Keep(fmt.Sprintf(`columns: ["%s"]`, column)).
		Group("").
		Distinct(fmt.Sprintf(`column: "%s"`, column)).
		String()
}

func (h *helper) genListingStmt() string {
	query := influx.Query{}
	query.Bucket("events").Range(h.genTimeDuration()).Measurement(h.eventType)
	query = h.addFilters(query)
	query.Pivot(convertValueToField).Group("").Sort(descByTime)
	if !h.isPageRequired() {
		return query.String()
	}

	offset := (h.Page.Number - 1) * h.Page.Size
	return query.
		Limit(fmt.Sprintf(`n: %d, offset: %d`, h.Page.Size, offset)).
		String()
}

func (h *helper) genAbstractStmt() string {
	query := influx.Query{}
	return query.Bucket("events").
		Range("start: 0").
		Measurement(h.eventType).
		Pivot(convertValueToField).
		Group("").
		Sort(descByTime).
		Limit(fmt.Sprintf(`n: %d`, h.limit)).
		String()
}

func (h *helper) genRankStmt() (string, error) {
	switch h.eventType {
	case "system":
		return h.genSystemRankStmt(), nil
	case "host":
		return h.genHostRankStmt(), nil
	case "instance":
		return h.genInstanceRankStmt(), nil
	}

	return "", fmt.Errorf("unsupported event type: %s", h.eventType)
}

func (h *helper) genSystemRankStmt() string {
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genTimeDuration()).
		Measurement("system").
		Filter(h.genFilter("category", h.category)).
		Filter(h.genFilter("severity", h.severity)).
		Group(`columns: ["key", "category", "severity"]`).
		Count(`column: "_value"`).
		Rename(`columns: {_value: "number"}`).
		Keep(`columns: ["key", "category", "severity", "number"]`).
		Group("").
		Sort(`columns: ["number"], desc: true`).
		Limit(fmt.Sprintf(`n: %d`, h.limit)).
		String()
}

func (h *helper) genHostRankStmt() string {
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genTimeDuration()).
		Measurement("host").
		Filter(h.genFilter("category", h.category)).
		Filter(h.genFilter("host", h.host)).
		Group(`columns: ["key", "category", "host"]`).
		Count(`column: "_value"`).
		Rename(`columns: {_value: "number"}`).
		Keep(`columns: ["key", "category", "host", "number"]`).
		Group("").
		Sort(`columns: ["number"], desc: true`).
		Limit(fmt.Sprintf(`n: %d`, h.limit)).
		String()
}

func (h *helper) genInstanceRankStmt() string {
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genTimeDuration()).
		Measurement("instance").
		Filter(h.genFilter("category", h.category)).
		Filter(h.genFilter("instance", h.instance)).
		Group(`columns: ["key", "category", "instance"]`).
		Count(`column: "_value"`).
		Rename(`columns: {_value: "number"}`).
		Keep(`columns: ["key", "category", "instance", "number"]`).
		Group("").
		Sort(`columns: ["number"], desc: true`).
		Limit(fmt.Sprintf(`n: %d`, h.limit)).
		String()
}

func (h *helper) genTimeDuration() string {
	if h.isPastRequired() {
		return fmt.Sprintf("start: %s", h.past)
	}

	return fmt.Sprintf(
		"start: %s, stop: %s",
		h.period.start,
		h.period.stop,
	)
}

func (h *helper) addFilters(query influx.Query) influx.Query {
	if h.isIdRequired() {
		query.Filter(h.genFilter("key", h.eventId))
	}

	if h.isCategoryRequired() {
		query.Filter(h.genFilter("category", h.category))
	}

	if h.isSeverityRequired() {
		query.Filter(h.genFilter("severity", h.severity))
	}

	if h.isHostRequired() {
		query.Filter(h.genFilter("host", h.host))
	}

	if h.isInstanceRequired() {
		query.Filter(h.genFuzzyFilter("_value", h.instance))
	}

	if h.isKeywordRequired() {
		query.Filter(h.genFuzzyFilter("_value", h.keyword))
	}

	return query
}

func (h *helper) genFilter(key, value string) string {
	return fmt.Sprintf(`fn: (r) => r.%s == "%s"`, key, value)
}

func (h *helper) genFuzzyFilter(key, value string) string {
	return fmt.Sprintf(`fn: (r) => r.%s =~ /%s/`, key, value)
}
