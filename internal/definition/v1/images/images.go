package images

import (
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module = "images"

	Db            = "images"
	ReqCollection = "requsets"

	GlanceDir              = "/mnt/cephfs/glance"
	DefaultOsDistro        = "os_distro"
	CubeDefinedOs          = "cubeDefinedOs"
	CubeDefinedDestination = "cubeDefinedDestination"
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

	reservedImages = []ReqOpts{
		{
			Name:                        "amphora-x64-haproxy",
			Os:                          "Ubuntu",
			Destination:                 "CubeStorage",
			Domain:                      "Default",
			Project:                     "admin",
			SourceFromAnotherHypervisor: false,
			Visibility:                  "private",
			Reserved: Reserved{
				Prefix: "amphora-",
				Type:   "lb",
			},
		},
		{
			Name:                        "manila-service-image",
			Os:                          "Ubuntu",
			Destination:                 "CubeStorage",
			Domain:                      "Default",
			Project:                     "admin",
			SourceFromAnotherHypervisor: false,
			Visibility:                  "private",
			Reserved: Reserved{
				Prefix: "manila-",
				Type:   "fs",
			},
		},
	}
)

type Image struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Os          string       `json:"os"`
	Destination string       `json:"destination"`
	Domain      string       `json:"domain"`
	Project     string       `json:"project"`
	Visibility  string       `json:"visibility"`
	SizeMiB     int64        `json:"sizeMiB"`
	CreatedAt   string       `json:"createdAt"`
	Status      status.Image `json:"status"`
}

type ReqOpts struct {
	Id                          string        `json:"id,omitempty" bson:"id"`
	File                        string        `json:"file,omitempty" bson:"file"`
	Name                        string        `json:"name" bson:"name"`
	Os                          string        `json:"os" bson:"os"`
	Destination                 string        `json:"destination" bson:"destination"`
	Domain                      string        `json:"domain" bson:"domain"`
	Project                     string        `json:"project" bson:"project"`
	SourceFromAnotherHypervisor bool          `json:"sourceFromAnotherHypervisor" bson:"sourceFromAnotherHypervisor"`
	Visibility                  string        `json:"visibility" bson:"visibility"`
	SizeMiB                     int64         `json:"sizeMiB" bson:"sizeMiB"`
	Status                      *status.Image `json:"status,omitempty" bson:"status,omitempty"`
	Reserved                    `json:"reserved" bson:"-"`
}

type CreateOpts struct {
	Dir            string       `json:"dir"`
	File           string       `json:"file"`
	Name           string       `json:"name"`
	Destination    string       `json:"destination"`
	Domain         string       `json:"domain"`
	Project        string       `json:"project"`
	PoolType       string       `json:"poolType"`
	AttributesType string       `json:"attributesType,omitempty"`
	Visibility     string       `json:"visibility"`
	StreamingLogs  chan float64 `json:"streaming,omitempty"`
	ReservedType   string       `json:"-"`
}

type Reserved struct {
	Prefix string `json:"prefix" bson:"-"`
	Type   string `json:"-" bson:"-"`
}

type Change struct {
	Id string
}

func (r *ReqOpts) GenCreateOpts() CreateOpts {
	poolType := "glance-images"
	visibility := r.Visibility
	if r.SourceFromAnotherHypervisor {
		poolType = "cinder-volumes"
		visibility = "public"
	}

	return CreateOpts{
		Dir:            GlanceDir,
		File:           r.File,
		Name:           r.Name,
		AttributesType: "default",
		Destination:    r.Destination,
		Domain:         r.Domain,
		PoolType:       poolType,
		Visibility:     visibility,
		StreamingLogs:  make(chan float64),
		ReservedType:   r.Reserved.Type,
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

func (r *ReqOpts) SetUploading() {
	if r.Status == nil {
		r.Status = &status.Image{}
	}

	r.Status.Current = status.Uploading
	r.Status.Desired = status.Uploaded
	r.Status.IsProcessing = true
	r.Status.ProcessPercent = 0
}

func (r *ReqOpts) SetImporting() {
	if r.Status == nil {
		r.Status = &status.Image{}
	}

	r.Status.Current = status.Importing
	r.Status.Desired = status.Imported
	r.Status.IsProcessing = true
	r.Status.ProcessPercent = 0
}

func (i *Image) GenSearchableObject() Image {
	return Image{
		Id:          search.NormalizeKeyword(i.Id),
		Name:        search.NormalizeKeyword(i.Name),
		Os:          search.NormalizeKeyword(i.Os),
		Destination: search.NormalizeKeyword(i.Destination),
		Domain:      search.NormalizeKeyword(i.Domain),
		Project:     search.NormalizeKeyword(i.Project),
		Visibility:  search.NormalizeKeyword(i.Visibility),
		Status: status.Image{
			Current:      search.NormalizeKeyword(i.Status.Current),
			IsProcessing: i.Status.IsProcessing,
		},
	}
}

func GetReserved() []ReqOpts {
	return reservedImages
}

func GetReservedInfo(name string) (*ReqOpts, bool) {
	for _, img := range reservedImages {
		if strings.HasPrefix(name, img.Reserved.Prefix) {
			return &img, true
		}
	}

	return nil, false
}
