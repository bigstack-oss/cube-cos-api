package firmwares

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module = "firmwares"

	TmpUploadDir       = "/tmp/firmwares"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir      = "/var/update"
	UpdateHistory  = "/var/appliance-db/update.history"
	UpdateProgress = "/var/run/cube-cos-api/progress.json"
)

type ReqOpts struct {
	Id          string          `json:"id"`
	Version     string          `json:"version"`
	PkgPath     string          `json:"pkgPath"`
	AutoRolling bool            `json:"autoRolling"`
	Status      status.Firmware `json:"status"`
}

type Firmware struct {
	Version      string          `json:"version" bson:"version"`
	ReleaseNotes string          `json:"releaseNotes" bson:"releaseNotes"`
	UpdatedAt    string          `json:"updatedAt" bson:"updatedAt"`
	Status       status.Firmware `json:"status" bson:"status"`
}

type Upadte struct {
	Current  string `yaml:"current"`
	Rollback string `yaml:"rollback"`
	History  []Raw  `yaml:"history"`
}

type Raw struct {
	Image     string `yaml:"image"`
	Type      string `yaml:"type"`
	Version   string `yaml:"version"`
	Variant   string `yaml:"variant"`
	BuiltAt   string `yaml:"built-at"`
	CreatedAt string `yaml:"created-at"`
}

type Upgrade struct {
	Version    string     `json:"version"`
	Progresses []Progress `json:"progresses"`
}

type Progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

func (u *ReqOpts) SetProcessing() {
	u.Status.Current = status.Upgrading
	u.Status.Desired = status.Upgraded
	u.Status.IsProcessing = true
}

func (u *ReqOpts) SetError() {
	u.Status.Desired = status.Error
	u.Status.IsProcessing = false
}

func (u *ReqOpts) SetCompleted() {
	u.Status.Desired = status.Updated
	u.Status.IsProcessing = false
}
