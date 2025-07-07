package triggers

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"

func (h *helper) GetAlertTypes(events []events.Event) ([]string, error) {
	alertTypes := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Type == "" {
			continue
		}

		_, exists := dedup[event.Type]
		if exists {
			continue
		}

		dedup[event.Type] = struct{}{}
		alertTypes = append(alertTypes, event.Type)
	}

	return alertTypes, nil
}

func (h *helper) GetSeverities(events []events.Event) ([]string, error) {
	severities := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Severity == "" {
			continue
		}

		_, exists := dedup[event.Severity]
		if exists {
			continue
		}

		dedup[event.Severity] = struct{}{}
		severities = append(severities, event.Severity)
	}

	return severities, nil
}

func (h *helper) GetCategories(events []events.Event) ([]string, error) {
	categories := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Category == "" {
			continue
		}

		_, exists := dedup[event.Category]
		if exists {
			continue
		}

		dedup[event.Category] = struct{}{}
		categories = append(categories, event.Category)
	}

	return categories, nil
}

func (h *helper) GetEventIds(events []events.Event) ([]string, error) {
	eventIds := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Id == "" {
			continue
		}

		_, exists := dedup[event.Id]
		if exists {
			continue
		}

		dedup[event.Id] = struct{}{}
		eventIds = append(eventIds, event.Id)
	}

	return eventIds, nil
}
