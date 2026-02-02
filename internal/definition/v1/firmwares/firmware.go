package firmwares

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module = "firmwares"

	Db               = "firmwares"
	UploadCollection = "upload"

	TmpUploadDir       = "/tmp/firmwares"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir           = "/var/update"
	UpdateHistory       = "/var/appliance-db/update.history"
	UpdateProgress      = "/var/lib/cube-cos-api/progress.json"
	ResolvedMarker      = "/var/lib/cube-cos-api/resolved"
	BootstrappingMarker = "/var/lib/cube-cos-api/bootstrapping"
	BootstrappingLog    = "/run/cube_bootstrap.log"
)

type ReqOpts struct {
	Id          string          `json:"id"`
	Hostname    string          `json:"hostname"`
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
	Version          string     `json:"version"`
	IsRollingApplied bool       `json:"isRollingApplied"`
	Progresses       []Progress `json:"progresses"`
}

type Progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

type BootstrappingStatus struct {
	Node   string `json:"node"`
	Return string `json:"return"`
	Stdout string `json:"stdout"`
}

type ResolvedStatus struct {
	HasFailureBeenResolved bool `json:"hasFailureBeenResolved"`
}

func (u *ReqOpts) SetInstalling() {
	u.Status.Current = status.Installing
	u.Status.Desired = status.Installed
	u.Status.IsProcessing = true
}

func (u *ReqOpts) SetError(err string) {
	u.Status.Current = status.Error
	u.Status.IsProcessing = false
	u.Status.Description = err
}

func (u *ReqOpts) SetWaitingReboot() {
	u.Status.Current = status.WaitingReboot
	u.Status.IsProcessing = false
}
