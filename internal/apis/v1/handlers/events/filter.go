package events

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
)

func (h *helper) addFilters(query influx.Query) influx.Query {
	if h.isIdRequired() {
		query.Filter(h.genFilter("key", h.eventId))
	}

	if h.isCategoriesRequired() {
		h.setCategoriesFilter(&query)
	}

	if h.isSeveritiesRequired() {
		h.setSeveritiesFilter(&query)
	}

	if h.isHostsRequired() {
		h.setHostsFilter(&query)
	}

	if h.isInstancesRequired() {
		h.setInstancesFilter(&query)
	}

	return query
}

func (h *helper) genFilter(key, value string) string {
	return fmt.Sprintf(`fn: (r) => r.%s == "%s"`, key, value)
}
