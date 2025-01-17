package v1

import "strings"

func SeverityFullName(severity string) string {
	switch strings.ToLower(severity) {
	case "c":
		return "Critical"
	case "w":
		return "Warning"
	case "i":
		return "Info"
	}

	return "Unknown"
}

type Event struct {
	Type        string                 `json:"type"`
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Host        string                 `json:"host"`
	Category    string                 `json:"category"`
	Service     string                 `json:"service"`
	Metadata    map[string]interface{} `json:"metadata"`
	Time        string                 `json:"time"`
}
