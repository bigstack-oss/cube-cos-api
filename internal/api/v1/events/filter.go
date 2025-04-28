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
