package licenses

import v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Licenses []v1.License `json:"licenses"`
	v1.Page  `json:"page"`
}
