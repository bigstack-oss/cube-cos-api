package health

import (
	"strings"

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

var (
	StatusOks = map[string]bool{
		status.Ok:            true,
		status.CheckDisabled: true,
		status.Disabled:      true,
		status.Checking:      true,
	}
	StatusOkDesciptions = []string{
		"checking returns the last result",
		"checking",
	}
)

type Report struct {
	Category string            `json:"category"`
	Service  string            `json:"service"`
	Status   status.Health     `json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"`
	Modules  []services.Module `json:"modules"`
}

type Check struct {
	Time     string `json:"time"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	*Error   `json:"error,omitempty"`
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

func (h *Check) IsOk() bool {
	_, found := StatusOks[h.Status]
	if found {
		return true
	}

	for _, desc := range StatusOkDesciptions {
		if strings.Contains(strings.ToLower(h.Status), desc) {
			return true
		}
	}

	return false
}

func (h *Check) IsFix() bool {
	return strings.Contains(h.Status, "fix")
}

func (h *Check) IsNg() bool {
	return !h.IsFix() && !h.IsOk()
}
