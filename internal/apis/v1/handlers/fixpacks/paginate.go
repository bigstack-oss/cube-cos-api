package fixpacks

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type fixpacksPage struct {
	Fixpacks   []fixpacks.Fixpack `json:"fixpacks"`
	pages.Page `json:"page"`
}

func (h *helper) paginateFixpacks(list []fixpacks.Fixpack) []fixpacks.Fixpack {
	if !h.page.IsRequired() {
		return list
	}

	left := min((h.page.Number-1)*h.page.Size, len(list))
	right := min(left+h.page.Size, len(list))
	return list[left:right]
}

func (h *helper) sortFixpacks(list *[]fixpacks.Fixpack) {
	sort.Slice(*list, func(i, j int) bool {
		return (*list)[i].UpdatedAt > (*list)[j].UpdatedAt
	})
}

func (h *helper) genPageInfo(list []fixpacks.Fixpack) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(list),
			TotalItemCount: int64(len(list)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(list)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(list)),
	}
}
