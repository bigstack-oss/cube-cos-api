package volumes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module = "volumes"

	Db                         = "volumes"
	ImageToVolumeReqCollection = "imageToVolumeRequsets"
)

type Volume struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	DiskTag    string        `json:"diskTag"`
	AttachedTo string        `json:"attachedTo"`
	Bootable   bool          `json:"bootable"`
	Shared     bool          `json:"shared"`
	SizeMiB    int64         `json:"sizeMiB"`
	CreatedAt  string        `json:"createdAt"`
	Status     status.Volume `json:"status"`
}

type Change struct {
	Id string
}

func (v *Volume) GenSearchableObject() Volume {
	return Volume{
		Id:         search.NormalizedKeyword(v.Id),
		Name:       search.NormalizedKeyword(v.Name),
		Type:       search.NormalizedKeyword(v.Type),
		DiskTag:    search.NormalizedKeyword(v.DiskTag),
		AttachedTo: search.NormalizedKeyword(v.AttachedTo),
		Bootable:   v.Bootable,
		Shared:     v.Shared,
		SizeMiB:    v.SizeMiB,
		CreatedAt:  search.NormalizedKeyword(v.CreatedAt),
		Status: status.Volume{
			Current:      search.NormalizedKeyword(v.Status.Current),
			IsProcessing: v.Status.IsProcessing,
		},
	}
}
