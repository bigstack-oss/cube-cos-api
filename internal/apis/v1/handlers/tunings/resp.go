package tunings

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type data struct {
	Tunings    []v1.Tuning `json:"tunings"`
	pages.Page `json:"page"`
}
