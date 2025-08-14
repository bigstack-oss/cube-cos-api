package firmwares

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module = "firmwares"

	TmpUploadDir       = "/tmp/firmwares"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir     = "/var/update"
	UpdateHistory = "/var/appliance-db/update.history"
)

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
