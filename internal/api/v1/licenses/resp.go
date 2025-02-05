package licenses

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Licenses        []definition.License `json:"licenses"`
	definition.Page `json:"page"`
}
