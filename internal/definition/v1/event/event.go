package event

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

type Options struct {
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

func (o *Options) GetSeverityFullName() string {
	return GetSeverityFullName(o.Severity)
}

func (o *Options) SetSearchIndex() {
	o.SearchIndex = uuid.New().String()
}

// note:
// in the current search lib(bleve), the algo is not able to detect the string if it include uppercase
// we've tried a few different init settings, but the result is not as expected as always
// currenlty, the only way we found is to convert all the string to lower case and inject to searcher
func (o *Options) GenSearchableObject() Options {
	return Options{
		SearchIndex: o.SearchIndex,
		Type:        strings.ToLower(o.Type),
		Id:          strings.ToLower(o.Id),
		Severity:    strings.ToLower(o.Severity),
		Description: strings.ToLower(o.Description),
		Host:        strings.ToLower(o.Host),
		Category:    strings.ToLower(o.Category),
		Service:     strings.ToLower(o.Service),
		Metadata:    maps.Clone(o.Metadata),
		Time:        o.Time,
	}
}
