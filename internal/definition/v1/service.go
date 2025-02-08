package v1

import "github.com/bigstack-oss/cube-cos-api/internal/status"

type Service struct {
	Name     string         `json:"service" bson:"service"`
	Category string         `json:"category" bson:"category"`
	Status   status.Details `json:"status,omitempty" bson:"status,omitempty"`
	Modules  []Module       `json:"modules" bson:"modules"`
}

type Module struct {
	Name             string         `json:"name" bson:"name"`
	Status           status.Details `json:"status,omitempty" bson:"status,omitempty"`
	IsAutoRepairable bool           `json:"-" bson:"isAutoRepairable"`
	Description      string         `json:"description,omitempty" bson:"description"`
}

func (s Service) CopyModuleEmptyStruct() Service {
	return Service{
		Name:     s.Name,
		Category: s.Category,
	}
}
