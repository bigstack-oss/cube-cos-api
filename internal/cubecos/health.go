package cubecos

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type Health struct {
	InUse  []definition.HealthInfo `json:"inUse"`
	Error  []definition.HealthInfo `json:"error"`
	Fixing []definition.HealthInfo `json:"fixing"`
}
