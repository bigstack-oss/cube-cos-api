package events

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"

func convertSystemSeverities(severities []string) []string {
	converted := []string{}
	for _, s := range severities {
		converted = append(
			converted,
			events.GetSeverityFullName(s),
		)
	}

	return converted
}
