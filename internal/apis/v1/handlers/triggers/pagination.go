package triggers

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (h *helper) sortTriggers(triggers *[]triggers.ApiSchema) {
	sort.Slice(*triggers, func(i, j int) bool {
		return (*triggers)[i].Name < (*triggers)[j].Name
	})
}

func (h *helper) paginateTriggers(triggerList []triggers.ApiSchema) []triggers.ApiSchema {
	if !h.page.IsRequired() {
		return triggerList
	}

	left := min((h.page.Number-1)*h.page.Size, len(triggerList))
	right := min(left+h.page.Size, len(triggerList))
	return triggerList[left:right]
}

func (h *helper) genPageInfo(triggerList []triggers.ApiSchema) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(triggerList),
			TotalItemCount: int64(len(triggerList)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(triggerList)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(triggerList)),
	}
}
