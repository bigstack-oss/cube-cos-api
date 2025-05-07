package licenses

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
)

type data struct {
	Licenses []license.Options `json:"licenses"`
	v1.Page  `json:"page"`
}
