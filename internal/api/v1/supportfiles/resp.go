package supportfiles

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

type fileSetList struct {
	SupportFileSet []support.FileSet `json:"supportFileSet"`
	v1.Page        `json:"page"`
}
