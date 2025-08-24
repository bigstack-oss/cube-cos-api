package fixpacks

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module        = "fixpacks"
	Db            = Module
	ReqCollection = "requests"

	TmpUploadDir       = "/tmp/fixpacks"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir = "/var/fixpack"
)

type ReqOpts struct {
	Hostname string         `json:"hostname" bson:"hostname"`
	Version  string         `json:"version" bson:"version"`
	Path     string         `json:"path" bson:"path"`
	Status   status.Fixpack `json:"status" bson:"status"`
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

func (r *ReqOpts) SetCompleted() {
	r.Status.Current = status.Completed
	r.Status.IsProcessing = false
}

func (r *ReqOpts) SetError() {
	r.Status.Current = status.Error
	r.Status.IsProcessing = false
}

func (r *ReqOpts) SetInstalling() {
	r.Status.Current = status.Installing
	r.Status.Desired = status.Installed
	r.Status.IsProcessing = true
}

func (r *ReqOpts) SetRollingBack() {
	r.Status.Current = status.RollingBack
	r.Status.Desired = status.Rollbacked
	r.Status.IsProcessing = true
}
