package licenses

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type licensePage struct {
	Licenses   []licenses.License `json:"licenses"`
	pages.Page `json:"page"`
}
