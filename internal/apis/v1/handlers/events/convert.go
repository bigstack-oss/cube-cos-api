package events

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
)

func convertSystemSeverities(severities []string) []string {
	converted := []string{}
	for _, s := range severities {
		converted = append(
			converted,
			event.GetSeverityFullName(s),
		)
	}

	return converted
}
