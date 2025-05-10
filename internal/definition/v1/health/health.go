package health

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/services"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module              = "healths"
	RepairingCollection = "repairing"
	repair              = "repair"

	AscSort  = `columns: ["_time"], desc: false`
	DescSort = `columns: ["_time"], desc: true`
)

type Report struct {
	Category string            `json:"category"`
	Service  string            `json:"service"`
	Status   status.Health     `json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"`
	Modules  []services.Module `json:"modules"`
}

type Check struct {
	Time   string `json:"time"`
	Status string `json:"status"`
	*Error `json:"error,omitempty"`
}

type Error struct {
	Type        string   `json:"type"`
	Reason      string   `json:"reason"`
	Nodes       []string `json:"nodes"`
	Description string   `json:"description"`
	Details     string   `json:"details"`
	Log         string   `json:"log"`
}

func RepairCollection() string {
	return repair
}

func (h *Check) IsNg() bool {
	return h.Status == status.Ng
}
