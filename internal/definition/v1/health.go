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
