package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

type filePage struct {
	SupportFileSet []support.FileSet `json:"supportFileSet"`
	pages.Page     `json:"page"`
}
