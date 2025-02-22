package tunings

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Tunings         []definition.Tuning `json:"tunings"`
	definition.Page `json:"page"`
}
