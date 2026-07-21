package events

import (
	"strings"

)

const (
	Module     = "event"
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
		"ids",
		"types",
		"category",
		"categories",
		"severity",
		"severities",
		"host",
		"hosts",
		"instance",
		"instances",
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
	Message     string         `json:"message"`
	Time        string         `json:"time"`
}

type Stat struct {
	Id           string  `json:"id"`
	Category     string  `json:"category"`
	Severity     string  `json:"severity,omitempty"`
	Host         string  `json:"host,omitempty"`
	InstanceId   string  `json:"instanceId,omitempty"`
	InstanceName string  `json:"instanceName,omitempty"`
	Percent      float64 `json:"percent"`
	Number       int64   `json:"number"`
}

type Filter struct {
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

func GetSeverityFullName(severity string) string {
	switch strings.ToLower(severity) {
	case "c":
		return "CRITICAL"
	case "w":
		return "WARNING"
	case "e":
		return "ERROR"
	case "i":
		return "INFO"
	}

	return severity
}

func GetSeverityShortName(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "C"
	case "warning":
		return "W"
	case "error":
		return "E"
	case "info":
		return "I"
	}

	return severity
}

func GetSeverityFullNames(severities []string) []string {
	names := []string{}
	for _, severity := range severities {
		names = append(
			names,
			GetSeverityShortName(severity),
		)
	}

	return names
}

func (e *Event) GetSeverityFullName() string {
	return GetSeverityFullName(e.Severity)
}

func (e *Event) SetCategory(metaObj map[string]any) {
	metaCategory, found := metaObj["category"]
	if !found {
		return
	}

	e.Category, _ = metaCategory.(string)
}

func (e *Event) SetService(metaObj map[string]any) {
	metaService, found := metaObj["service"]
	if !found {
		return
	}

	e.Service, _ = metaService.(string)
}

func (e *Event) SetHostname(metaObj map[string]any) {
	metaHost, found := metaObj["host"]
	if !found {
		return
	}

	e.Host, _ = metaHost.(string)
}
