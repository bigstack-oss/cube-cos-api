package healths

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
)

const (
	healthMeasurement   = `fn: (r) => r._measurement == "health"`
	convertValueToField = `rowKey: ["_time","component","node","code"], columnKey: ["_field"], valueColumn: "_value"`
	AscSort             = `columns: ["_time"], desc: false`
	repairingCode       = 0
)

func (h *helper) genModuleHealthHistoryQuery(onlyLast bool) string {
	query := influx.Query{}
	query.Bucket("events").
		Range(h.genTimeDuration()).
		Filter(healthMeasurement).
		Filter(h.genModuleFilter()).
		Pivot(convertValueToField).
		Group("").
		Sort(AscSort)

	if onlyLast {
		query.Limit("n: 1")
	}

	return query.String()
}

func (h *helper) genTimeDuration() string {
	if queries.ArePeriodAndPastEmpty(h.c) {
		return "start: -24h"
	}

	if queries.IsPastRequired(h.c) {
		return fmt.Sprintf("start: -%s", h.past)
	}

	return fmt.Sprintf(
		"start: %s, stop: %s",
		h.period.Start,
		h.period.Stop,
	)
}

func (h *helper) genModuleFilter() string {
	return fmt.Sprintf(`fn: (r) => r.component == "%s"`, h.moduleType)
}
