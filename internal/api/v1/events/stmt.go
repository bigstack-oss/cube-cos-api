package events

var (
	eventQueryTemplate = `
		from(bucket: "events")
			|> range(start: %q, stop: %q)
			|> filter(fn: (r) => r._measurement == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> sort(columns: ["_time"], desc: true)
	`
)
