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

	eventNonPagingQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
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
)

func (h *helper) genCountQueryStmt() string {
	return fmt.Sprintf(
		eventCountQueryTemplate,
		h.period.start,
		h.period.stop,
		h.eventType,
	)
}

func (h *helper) genQueryStmt() string {
	if !h.isPaginationEnabled() {
		return h.genNonPagingQueryStmt()
	}

	return h.genPagingQueryStmt()
}

func (h *helper) genNonPagingQueryStmt() string {
	return fmt.Sprintf(
		eventNonPagingQueryTemplate,
		h.period.start,
		h.period.stop,
		h.eventType,
	)
}

func (h *helper) genPagingQueryStmt() string {
	offset := (h.page.Number - 1) * h.page.Size
	return fmt.Sprintf(
		eventPagingQueryTemplate,
		h.period.start,
		h.period.stop,
		h.eventType,
		h.page.Size,
		offset,
	)
}
