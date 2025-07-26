package images

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module = "images"

	Db            = "images"
	ReqCollection = "requsets"

	GlanceDir = "/mnt/cephfs/glance/images"
)

var (
	Visibilitise = []string{"public", "private"}
	Destinations = []string{"CubeStorage"}
	Oses         = []string{
		"CentOS",
		"Fedora",
		"Ubuntu",
		"Debian",
		"Windows",
		"Rocky",
		"FreeBSD",
		"CoreOS",
		"Arch",
		"Others",
	}
)

type ReqOpts struct {
	Id                          string        `json:"id,omitempty" bson:"id"`
	File                        string        `json:"file" bson:"file"`
	Name                        string        `json:"name" bson:"name"`
	Os                          string        `json:"os" bson:"os"`
	Destination                 string        `json:"destination" bson:"destination"`
	Domain                      string        `json:"domain" bson:"domain"`
	SourceFromAnotherHypervisor bool          `json:"sourceFromAnotherHypervisor" bson:"sourceFromAnotherHypervisor"`
	Visibility                  string        `json:"visibility" bson:"visibility"`
	Status                      *status.Image `json:"status,omitempty" bson:"status,omitempty"`
}

type CreateOpts struct {
	Dir         string `json:"dir"`
	File        string `json:"file"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
	Domain      string `json:"domain"`
	PoolType    string `json:"poolType"`
	Visibility  string `json:"visibility"`
}

func (r *ReqOpts) GenCreateOpts() CreateOpts {
	poolType := "glance-images"
	if r.SourceFromAnotherHypervisor {
		poolType = "cinder-volumes"
	}

	return CreateOpts{
		Dir:         GlanceDir,
		File:        r.File,
		Name:        r.Name,
		Destination: r.Destination,
		Domain:      r.Domain,
		PoolType:    poolType,
		Visibility:  r.Visibility,
	}
}

func (r *ReqOpts) SetCompleted() {
	if r.Status == nil {
		r.Status = &status.Image{}
	}

	r.Status.Current = status.Completed
	r.Status.IsProcessing = false
}

func (r *ReqOpts) SetError() {
	if r.Status == nil {
		r.Status = &status.Image{}
	}

	r.Status.Current = status.Error
	r.Status.IsProcessing = false
}
