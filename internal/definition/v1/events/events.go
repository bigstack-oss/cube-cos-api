package events

import "strings"

const (
	Name       = "events"
	TimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

var (
	supportEventTypes = map[string]bool{
		"system":   true,
		"host":     true,
		"instance": true,
	}

	filterConditions = []string{
		"id",
		"category",
		"severity",
		"host",
		"instance",
		"keyword",
	}
)

type Event struct {
	Type        string         `json:"type"`
	Severity    string         `json:"severity"`
	Id          string         `json:"id"`
	Description string         `json:"description"`
	Host        string         `json:"host"`
	Category    string         `json:"category"`
	Service     string         `json:"service"`
	Metadata    map[string]any `json:"metadata"`
	Time        string         `json:"time"`
}

type EventStat struct {
	Id      string  `json:"id"`
	Percent float64 `json:"percent"`
	Number  int64   `json:"number"`
	Query   string  `json:"query"`
}

type EventFilter struct {
	System   SystemFilter   `json:"system"`
	Instance InstanceFilter `json:"instance"`
	Host     HostFilter     `json:"host"`
}

type SystemFilter struct {
	Severities []string `json:"severities"`
	Categories []string `json:"categories"`
}

type InstanceFilter struct {
	Ids        []string `json:"ids"`
	Categories []string `json:"categories"`
}

type HostFilter struct {
	Names      []string `json:"names"`
	Categories []string `json:"categories"`
}

func IsValidType(t string) bool {
	return supportEventTypes[t]
}

func GetFilterConditions() []string {
	return filterConditions
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

	return severity
}

func SeverityShortName(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "C"
	case "warning":
		return "W"
	case "info":
		return "I"
	}

	return severity
}
