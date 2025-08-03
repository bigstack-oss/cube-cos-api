package volumes

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
)

type volumePage struct {
	Volumes    []volumes.Volume `json:"volumes"`
	pages.Page `json:"page"`
}

func (h *helper) paginateVolumes(volumes []volumes.Volume) []volumes.Volume {
	if !h.page.IsRequired() {
		return volumes
	}

	left := min((h.page.Number-1)*h.page.Size, len(volumes))
	right := min(left+h.page.Size, len(volumes))
	return volumes[left:right]
}

func (h *helper) sortVolumes(volumes *[]volumes.Volume) {
	sort.Slice(*volumes, func(i, j int) bool {
		return (*volumes)[i].CreatedAt > (*volumes)[j].CreatedAt
	})
}

func (h *helper) genPageInfo(volumes []volumes.Volume) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(volumes),
			TotalItemCount: int64(len(volumes)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(volumes)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(volumes)),
	}
}
