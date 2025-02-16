package events

import (
	"fmt"
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

	eventNonPagingQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> group()
			|> sort(columns: ["_time"], desc: true)
	`

	eventIdNonPagingQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r.key == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> group()
			|> sort(columns: ["_time"], desc: true)
	`

	eventPagingQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> group()
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d, offset: %d)
	`

	eventIdPagingQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r.key == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> group()
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d, offset: %d)
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
)

func (h *helper) genCountQueryStmt() string {
	if h.isIdRequired() {
		return fmt.Sprintf(
			eventIdCountQueryTemplate,
			h.period.start,
			h.period.stop,
			h.eventType,
			h.eventId,
		)
	}

	return fmt.Sprintf(
		eventCountQueryTemplate,
		h.period.start,
		h.period.stop,
		h.eventType,
	)
}

func (h *helper) genListingStmt() string {
	if !h.isPageRequired() {
		return h.genNonPagingQueryStmt()
	}

	return h.genPagingQueryStmt()
}

func (h *helper) genNonPagingQueryStmt() string {
	if h.isIdRequired() {
		return fmt.Sprintf(
			eventIdNonPagingQueryTemplate,
			h.period.start,
			h.period.stop,
			h.eventType,
			h.eventId,
		)
	}

	return fmt.Sprintf(
		eventNonPagingQueryTemplate,
		h.period.start,
		h.period.stop,
		h.eventType,
	)
}

func (h *helper) genPagingQueryStmt() string {
	offset := (h.Page.Number - 1) * h.Page.Size
	if h.isIdRequired() {
		return fmt.Sprintf(
			eventIdPagingQueryTemplate,
			h.period.start,
			h.period.stop,
			h.eventType,
			h.eventId,
			h.Page.Size,
			offset,
		)
	}

	return fmt.Sprintf(
		eventPagingQueryTemplate,
		h.period.start,
		h.period.stop,
		h.eventType,
		h.Page.Size,
		offset,
	)
}

func (h *helper) genAbstractStmt() string {
	return fmt.Sprintf(
		eventLimitingQueryTemplate,
		h.eventType,
		h.limit,
	)
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
	return fmt.Sprintf(
		eventSystemRankQueryTemplate,
		h.period.start,
		h.period.stop,
		h.category,
		h.severity,
		h.limit,
	)
}

func (h *helper) genHostRankStmt() string {
	return fmt.Sprintf(
		eventHostRankQueryTemplate,
		h.period.start,
		h.period.stop,
		h.category,
		h.host,
		h.limit,
	)
}

func (h *helper) genInstanceRankStmt() string {
	return fmt.Sprintf(
		eventInstanceRankQueryTemplate,
		h.period.start,
		h.period.stop,
		h.category,
		h.instance,
		h.limit,
	)
}
