package events

import v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Events    any `json:"events"`
	*v1.Page  `json:"page,omitempty"`
	*v1.Limit `json:"limit,omitempty"`
}
