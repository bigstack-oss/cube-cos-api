package v1

import (
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	Healths = "healths"
	repair  = "repair"
)

type Health struct {
	Category string         `json:"category"`
	Service  string         `json:"service"`
	Status   status.Details `json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"`
	Modules  []Module       `json:"modules"`
}

func HealthDB() string {
	return Healths
}

func RepairCollection() string {
	return repair
}

type HealthCheck struct {
	Time        string `json:"time"`
	Code        int    `json:"code"`
	Component   string `json:"component"`
	Description string `json:"description"`
	Details     string `json:"details"`
	Log         string `json:"log"`
	Node        string `json:"node"`
	Status      string `json:"status"`
	*Error      `json:"error,omitempty"`
}

type Error struct {
	Type        string   `json:"type"`
	Reason      string   `json:"reason"`
	Nodes       []string `json:"nodes"`
	Description string   `json:"description"`
	Details     string   `json:"details"`
	Log         string   `json:"log"`
}
