package triggers

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"

func (h *helper) GetAlertTypes(events []events.Event) ([]string, error) {
	alertTypes := []string{}
	for _, event := range events {
		if event.Type != "" {
			alertTypes = append(alertTypes, event.Type)
		}
	}

	return alertTypes, nil
}

func (h *helper) GetSeverities(events []events.Event) ([]string, error) {
	severities := []string{}
	for _, event := range events {
		if event.Severity != "" {
			severities = append(severities, event.Severity)
		}
	}

	return severities, nil
}

func (h *helper) GetCategories(events []events.Event) ([]string, error) {
	categories := []string{}
	for _, event := range events {
		if event.Category != "" {
			categories = append(categories, event.Category)
		}
	}

	return categories, nil
}

func (h *helper) GetEventIds(events []events.Event) ([]string, error) {
	eventIds := []string{}
	for _, event := range events {
		if event.Id != "" {
			eventIds = append(eventIds, event.Id)
		}
	}

	return eventIds, nil
}
