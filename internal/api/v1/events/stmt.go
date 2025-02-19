package events

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
)

var (
	eventCountQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r._field == "message")
			|> count()
	`

	eventIdCountQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r.key == "%s")
			|> count()
	`

	eventLimitingQueryTemplate = `
		from(bucket: "events")
			|> range(start: 0)
			|> filter(fn: (r) => r._measurement == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> group()
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d)
	`

	eventSystemRankQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "system")
			|> filter(fn: (r) => r.category == "%s")
			|> filter(fn: (r) => r.severity == "%s")
			|> group(columns: ["key", "category", "severity"])
			|> count(column: "_value")
			|> rename(columns: {_value: "number"})
			|> keep(columns: ["key", "category", "severity", "number"])
			|> group()
			|> sort(columns: ["number"], desc: true)
			|> limit(n: %d)
	`

	eventHostRankQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "host")
			|> filter(fn: (r) => r.category == "%s")
			|> filter(fn: (r) => r.host == "%s")
			|> group(columns: ["key", "category", "host"])
			|> count(column: "_value")
			|> rename(columns: {_value: "number"})
			|> keep(columns: ["key", "category", "host", "number"])
			|> group()
			|> sort(columns: ["number"], desc: true)
			|> limit(n: %d)
	`

	eventInstanceRankQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "instance")
			|> filter(fn: (r) => r.category == "%s")
			|> filter(fn: (r) => r.instance == "%s")
			|> group(columns: ["key", "category", "instance"])
			|> count(column: "_value")
			|> rename(columns: {_value: "number"})
			|> keep(columns: ["key", "category", "instance", "number"])
			|> group()
			|> sort(columns: ["number"], desc: true)
			|> limit(n: %d)
	`

	eventFilterConditionQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> keep(columns: ["%s"])
			|> group()                    
			|> distinct(column: "%s")
	`
)

func (h *helper) genCountQueryStmt() string {
	// if h.isIdRequired() {
	// 	return fmt.Sprintf(
	// 		eventIdCountQueryTemplate,
	// 		h.period.start,
	// 		h.period.stop,
	// 		h.eventType,
	// 		h.eventId,
	// 	)
	// }

	// return fmt.Sprintf(
	// 	eventCountQueryTemplate,
	// 	h.period.start,
	// 	h.period.stop,
	// 	h.eventType,
	// )

	query := influx.Query{}
	query.Bucket("events").
		Range(h.genStartStopRange()).
		Measurement(h.eventType)

	h.addFilters(&query)
	return query.Count("").String()
}

func (h *helper) genListingStmt() string {
	query := influx.Query{}
	query.Bucket("events").
		Range(h.genStartStopRange()).
		Measurement(h.eventType)

	h.addFilters(&query)

	query.
		Pivot(`rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value"`).
		Group("").
		Sort(`columns: ["_time"], desc: true`)

	if !h.isPageRequired() {
		return query.String()
	}

	offset := (h.Page.Number - 1) * h.Page.Size
	return query.
		Limit(fmt.Sprintf(`n: %d, offset: %d`, h.Page.Size, offset)).
		String()
}

func (h *helper) addFilters(query *influx.Query) {
	if h.isIdRequired() {
		query.Filter(fmt.Sprintf(`fn: (r) => r.key == "%s"`, h.eventId))
	}

	if h.isCategoryRequired() {
		query.Filter(fmt.Sprintf(`fn: (r) => r.category == "%s"`, h.category))
	}

	if h.isSeverityRequired() {
		query.Filter(fmt.Sprintf(`fn: (r) => r.severity == "%s"`, h.severity))
	}

	if h.isHostRequired() {
		query.Filter(fmt.Sprintf(`fn: (r) => r.host == "%s"`, h.host))
	}

	if h.isInstanceRequired() {
		query.Filter(fmt.Sprintf(`fn: (r) => r._value =~ /%s/`, h.instance))
	}

	if h.isKeywordRequired() {
		query.Filter(fmt.Sprintf(`fn: (r) => r._value =~ /%s/`, h.keyword))
	}
}

func (h *helper) genAbstractStmt() string {
	query := influx.Query{}
	return query.Bucket("events").
		Range("start: 0").
		Measurement(h.eventType).
		Pivot(`rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value"`).
		Group("").
		Sort(`columns: ["_time"], desc: true`).
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
	// return fmt.Sprintf(
	// 	eventSystemRankQueryTemplate,
	// 	h.period.start,
	// 	h.period.stop,
	// 	h.category,
	// 	h.severity,
	// 	h.limit,
	// )
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genStartStopRange()).
		Measurement("system").
		Filter(fmt.Sprintf(`fn: (r) => r.category == "%s"`, h.category)).
		Filter(fmt.Sprintf(`fn: (r) => r.severity == "%s"`, h.severity)).
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
	// return fmt.Sprintf(
	// 	eventHostRankQueryTemplate,
	// 	h.period.start,
	// 	h.period.stop,
	// 	h.category,
	// 	h.host,
	// 	h.limit,
	// )
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genStartStopRange()).
		Measurement("host").
		Filter(fmt.Sprintf(`fn: (r) => r.category == "%s"`, h.category)).
		Filter(fmt.Sprintf(`fn: (r) => r.host == "%s"`, h.host)).
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
	// return fmt.Sprintf(
	// 	eventInstanceRankQueryTemplate,
	// 	h.period.start,
	// 	h.period.stop,
	// 	h.category,
	// 	h.instance,
	// 	h.limit,
	// )
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genStartStopRange()).
		Measurement("instance").
		Filter(fmt.Sprintf(`fn: (r) => r.category == "%s"`, h.category)).
		Filter(fmt.Sprintf(`fn: (r) => r.instance == "%s"`, h.instance)).
		Group(`columns: ["key", "category", "instance"]`).
		Count(`column: "_value"`).
		Rename(`columns: {_value: "number"}`).
		Keep(`columns: ["key", "category", "instance", "number"]`).
		Group("").
		Sort(`columns: ["number"], desc: true`).
		Limit(fmt.Sprintf(`n: %d`, h.limit)).
		String()
}

func (h *helper) genFilterConditionStmt(eventType, column string) string {
	// return fmt.Sprintf(
	// 	eventFilterConditionQueryTemplate,
	// 	h.period.start,
	// 	h.period.stop,
	// 	eventType,
	// 	column,
	// 	column,
	// )
	query := influx.Query{}
	return query.Bucket("events").
		Range(h.genStartStopRange()).
		Measurement(eventType).
		Keep(fmt.Sprintf(`columns: ["%s"]`, column)).
		Group("").
		Distinct(fmt.Sprintf(`column: "%s"`, column)).
		String()
}

func (h *helper) genStartStopRange() string {
	return fmt.Sprintf(
		"start: %s, stop: %s",
		h.period.start,
		h.period.stop,
	)
}
