package firmwares

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type firmwarePage struct {
	Firmwares  []firmwares.Firmware `json:"firmwares"`
	pages.Page `json:"page"`
}

func (h *helper) paginateFirmwares(list []firmwares.Firmware) []firmwares.Firmware {
	if !h.page.IsRequired() {
		return list
	}

	left := min((h.page.Number-1)*h.page.Size, len(list))
	right := min(left+h.page.Size, len(list))
	return list[left:right]
}

func (h *helper) sortFirmwares(list *[]firmwares.Firmware) {
	sort.Slice(*list, func(i, j int) bool {
		return (*list)[i].UpdatedAt > (*list)[j].UpdatedAt
	})
}

func (h *helper) genPageInfo(list []firmwares.Firmware) pages.Page {
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
