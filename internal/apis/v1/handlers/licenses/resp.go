package licenses

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type licensePages struct {
	Licenses   []license.Options `json:"licenses"`
	pages.Page `json:"page"`
}
