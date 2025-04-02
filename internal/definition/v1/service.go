package v1

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

type Service struct {
	Name               string          `json:"name" bson:"name"`
	Category           string          `json:"category" bson:"category"`
	Status             *status.Details `json:"status,omitempty" bson:"status,omitempty"`
	Modules            []Module        `json:"modules" bson:"modules"`
	IsInternalViewOnly bool            `json:"-" bson:"isInternalViewOnly"`
}

type Module struct {
	Name         string          `json:"name" bson:"name"`
	Status       *status.Details `json:"status,omitempty" bson:"status,omitempty"`
	IsRepairable bool            `json:"-" bson:"isRepairable"`
	Description  string          `json:"description,omitzero" bson:"description"`
}

type ReairingInfo struct {
	Node    string `json:"node"`
	Fixing  string `json:"fixing"`
	Service string `json:"service"`
	Pid     string `json:"pid"`
	Elaps   string `json:"elaps"`
}

func (s *Service) IsStatusOk() bool {
	return s.Status.Current != status.Ok
}

func (s *Service) SetRepairingStatus(repairingInfo ReairingInfo) {
	s.Status = &status.Details{
		Current:  status.Repairing,
		IsFixing: true,
		Description: fmt.Sprintf(
			"module %s is repairing(elaps %s)",
			repairingInfo.Service,
			repairingInfo.Elaps,
		),
	}
}

func (s *Service) CopyModuleEmptyStruct() Service {
	return Service{
		Name:     s.Name,
		Category: s.Category,
	}
}

func (s *Service) ConvergeUnhealthyStatus(moduleName string, unhealthy *HealthCheck) {
	if s.Status == nil {
		s.Status = &status.Details{
			Current:     status.Ng,
			Description: "failure modules detected: ",
		}
	}

	s.Status.Description += fmt.Sprintf(
		"%s(%s)",
		moduleName,
		unhealthy.Description,
	)
}

func (m *Module) InitOkStatus() {
	m.Status = status.NewOk()
}

func (m *Module) SetUnhealthyStatus(unhealthy *HealthCheck) {
	m.Status = &status.Details{
		Current:     unhealthy.Status,
		Description: unhealthy.Description,
	}
}

func (m *Module) SetRepairingStatus() {
	m.Status = &status.Details{
		Current:  status.Repairing,
		IsFixing: true,
	}
}
