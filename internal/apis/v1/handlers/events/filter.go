package events

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
)

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
		query.Filter(h.genFilter("instance", h.instance))
	}

	return query
}

func (h *helper) genFilter(key, value string) string {
	return fmt.Sprintf(`fn: (r) => r.%s == "%s"`, key, value)
}
