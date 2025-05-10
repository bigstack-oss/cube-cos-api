package tunings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
)

type data struct {
	Tunings    []tunings.Tuning `json:"tunings"`
	pages.Page `json:"page"`
}
