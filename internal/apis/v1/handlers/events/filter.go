package events

import (
	"fmt"
	"strings"

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

func (h *helper) filteredPredefinedEvents(list []predefinedEvent) []predefinedEvent {
	if h.isTypesRequired() {
		list = h.filterByTypes(list)
	}

	if h.isCategoriesRequired() {
		list = h.filterByCategories(list)
	}

	if h.isSeveritiesRequired() {
		list = h.filterBySeverities(list)
	}

	if h.isEventIdsRequired() {
		list = h.filterByEventIds(list)
	}

	return list
}

func (h *helper) filterByTypes(list []predefinedEvent) []predefinedEvent {
	filtered := []predefinedEvent{}
	for _, event := range list {
		for _, eventType := range h.eventTypes {
			if strings.EqualFold(event.Type, eventType) {
				filtered = append(filtered, event)
				break
			}
		}
	}

	return filtered
}

func (h *helper) filterByCategories(list []predefinedEvent) []predefinedEvent {
	filtered := []predefinedEvent{}
	for _, event := range list {
		for _, category := range h.categories {
			if strings.EqualFold(event.Category, category) {
				filtered = append(filtered, event)
				break
			}
		}
	}

	return filtered
}

func (h *helper) filterBySeverities(list []predefinedEvent) []predefinedEvent {
	filtered := []predefinedEvent{}
	for _, event := range list {
		for _, severity := range h.severities {
			if strings.EqualFold(event.Severity, severity) {
				filtered = append(filtered, event)
				break
			}
		}
	}

	return filtered
}

func (h *helper) filterByEventIds(list []predefinedEvent) []predefinedEvent {
	filtered := []predefinedEvent{}
	for _, event := range list {
		for _, eventId := range h.eventIds {
			if strings.EqualFold(event.Id, eventId) {
				filtered = append(filtered, event)
				break
			}
		}
	}

	return filtered
}
