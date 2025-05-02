package events

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
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
	return query.Pivot(convertValueToField).
		Group("").
		String()
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
	return query.
		Pivot(convertValueToField).
		Group("").
		Sort(descByTime).
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
	query.
		Bucket("events").
		Range(h.genTimeDuration()).
		Measurement("system").
		Filter(`fn: (r) => r._field == "message"`)

	if h.category != "" {
		query.Filter(h.genFilter("category", h.category))
	}
	if h.severity != "" {
		query.Filter(h.genFilter("severity", h.severity))
	}

	return query.
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
	query.
		Bucket("events").
		Range(h.genTimeDuration()).
		Measurement("host").
		Filter(`fn: (r) => r._field == "message"`)

	if h.category != "" {
		query.Filter(h.genFilter("category", h.category))
	}
	if h.host != "" {
		query.Filter(h.genFilter("host", h.host))
	}

	return query.
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
	query.
		Bucket("events").
		Range(h.genTimeDuration()).
		Measurement("instance").
		Filter(`fn: (r) => r._field == "message"`)

	if h.category != "" {
		query.Filter(h.genFilter("category", h.category))
	}
	if h.instance != "" {
		query.Filter(h.genFilter("instance", h.instance))
	}

	return query.
		Group(`columns: ["key", "category", "instance", "vm_name"]`).
		Count(`column: "_value"`).
		Rename(`columns: {_value: "number"}`).
		Keep(`columns: ["key", "category", "instance", "vm_name", "number"]`).
		Group("").
		Sort(`columns: ["number"], desc: true`).
		Limit(fmt.Sprintf(`n: %d`, h.limit)).
		String()
}

func (h *helper) genTimeDuration() string {
	if query.IsPastRequired(h.c) {
		return fmt.Sprintf("start: -%s", h.past)
	}

	return fmt.Sprintf(
		"start: %s, stop: %s",
		h.Period.Start,
		h.Period.Stop,
	)
}
