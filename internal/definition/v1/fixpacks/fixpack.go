package fixpacks

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module = "fixpacks"

	TmpUploadDir       = "/tmp/fixpacks"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir = "/var/fixpack"
)

type InstallReqOpts struct {
}

type Raw struct {
	Id                 string   `json:"id" bson:"id"`
	Name               string   `json:"name" bson:"name"`
	SupportedFirmwares []string `json:"supportedFirmwares" bson:"supportedFirmwares"`
	Description        string   `json:"description" bson:"description"`
	Details            string   `json:"details"`
}

type Fixpack struct {
	Version        string         `json:"version"`
	Name           string         `json:"name"`
	Note           string         `json:"note"`
	Details        string         `json:"details"`
	UpdatedAt      string         `json:"updatedAt"`
	RebootRequired bool           `json:"rebootRequired"`
	Status         status.Fixpack `json:"status"`
}
