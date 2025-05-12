package events

import (
	"maps"
	"strings"

	"github.com/google/uuid"
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
	SearchIndex string         `json:"-"`
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
		return "Critical"
	case "w":
		return "Warning"
	case "e":
		return "Error"
	case "i":
		return "Info"
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

func (e *Event) SetSearchIndex() {
	e.SearchIndex = uuid.New().String()
}

// note:
// in the current search lib(bleve), the algo is not able to detect the string if it include uppercase
// we've tried a few different init settings, but the result is not as expected as always
// currenlty, the only way we found is to convert all the string to lower case and inject to searcher
func (e *Event) GenSearchableObject() Event {
	return Event{
		SearchIndex: e.SearchIndex,
		Type:        strings.ToLower(e.Type),
		Id:          strings.ToLower(e.Id),
		Severity:    strings.ToLower(e.Severity),
		Description: strings.ToLower(e.Description),
		Host:        strings.ToLower(e.Host),
		Category:    strings.ToLower(e.Category),
		Service:     strings.ToLower(e.Service),
		Metadata:    maps.Clone(e.Metadata),
		Time:        e.Time,
	}
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
