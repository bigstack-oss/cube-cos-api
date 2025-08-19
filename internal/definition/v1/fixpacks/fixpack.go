package fixpacks

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module    = "fixpacks"
	UpdateDir = "/var/fixpack"
)

type InstallReqOpts struct {
}

type Fixpack struct {
	Version        string         `json:"version"`
	Note           string         `json:"note"`
	UpdatedAt      string         `json:"updatedAt"`
	RebootRequired bool           `json:"rebootRequired"`
	Status         status.Fixpack `json:"status"`
}
