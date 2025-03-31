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

func (s *Service) CopyModuleEmptyStruct() Service {
	return Service{
		Name:     s.Name,
		Category: s.Category,
	}
}

func (s *Service) SetErr(health *HealthCheck) {
	if s.Status == nil {
		s.Status = &status.Details{
			Current:     status.Ng,
			Description: "failure modules detected: ",
		}
	}

	s.Status.Description += fmt.Sprintf(
		"%s(%s)",
		health.Component,
		health.Description,
	)
}

func (m *Module) SetErr(health *HealthCheck) {
	m.Status = &status.Details{
		Current:     health.Status,
		Description: health.Description,
	}
}
