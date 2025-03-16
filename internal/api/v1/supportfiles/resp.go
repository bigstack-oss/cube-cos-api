package supportfiles

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

type data struct {
	SupportFiles []v1.SupportFile `json:"supportFiles"`
	v1.Page      `json:"page"`
}
