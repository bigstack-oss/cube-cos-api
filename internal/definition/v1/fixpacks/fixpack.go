package fixpacks

import (
	"regexp"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module           = "fixpacks"
	Db               = Module
	ReqCollection    = "requests"
	UploadCollection = "upload"

	TmpUploadDir       = "/tmp/fixpacks"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir        = "/var/fixpack"
	RollbackDir      = "/var/fixpack_rollback"
	NeedRebootMarker = "/run/need_reboot"
	Info             = "fixpack.info"
)

var (
	RollbackFileRegex = regexp.MustCompile(`^fixpack-(\d+)$`)
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
	RebootRequired     []string `json:"rebootRequired" bson:"rebootRequired"`
	Rollbackable       bool     `json:"rollbackable" bson:"rollbackable"`
	Description        string   `json:"description" bson:"description"`
	Details            string   `json:"details"`
}

type Fixpack struct {
	Version        string         `json:"version"`
	Action         string         `json:"-"`
	Name           string         `json:"name"`
	Note           string         `json:"note"`
	Details        string         `json:"details"`
	UpdatedAt      string         `json:"updatedAt"`
	RebootRequired bool           `json:"rebootRequired"`
	TargetNodes    []string       `json:"-"`
	Status         status.Fixpack `json:"status"`
}

func (r *ReqOpts) SetCompleted() {
	switch r.Status.Desired {
	case status.Installed:
		r.Status.Current = status.Installed
	case status.Rollbacked:
		r.Status.Current = status.Rollbacked
	default:
		r.Status.Current = status.Unknown
	}

	r.Status.IsProcessing = false
}

func (r *ReqOpts) SetFailed() {
	r.Status.Current = status.Failed
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
