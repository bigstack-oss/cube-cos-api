package v1

import "strings"

const (
	Events = "events"
)

type Event struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Id          string                 `json:"id"`
	Description string                 `json:"description"`
	Host        string                 `json:"host"`
	Category    string                 `json:"category"`
	Service     string                 `json:"service"`
	Metadata    map[string]interface{} `json:"metadata"`
	Time        string                 `json:"time"`
}

type EventStat struct {
	Id      string  `json:"id"`
	Percent float64 `json:"percent"`
	Number  int64   `json:"number"`
	Query   string  `json:"query"`
}

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
