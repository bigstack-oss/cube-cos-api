package events

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Events []definition.Event `json:"events"`
	Page   page               `json:"page"`
}
