package events

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Events            interface{} `json:"events"`
	*definition.Page  `json:"page,omitempty"`
	*definition.Limit `json:"limit,omitempty"`
}
