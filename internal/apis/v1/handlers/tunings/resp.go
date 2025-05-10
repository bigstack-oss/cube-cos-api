package tunings

import v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Tunings []v1.Tuning `json:"tunings"`
	v1.Page `json:"page"`
}
