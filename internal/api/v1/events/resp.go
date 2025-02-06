package events

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type resp struct {
	Events          []definition.Event `json:"events"`
	definition.Page `json:"page"`
}
