package tunings

import (
	"math"
	"sort"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateTunings(tunings []v1.Tuning) ([]v1.Tuning, error) {
	if !h.Page.IsRequired() {
		return tunings, nil
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(tunings))
	right := min(left+h.Page.Size, len(tunings))
	return tunings[left:right], nil
}

func (h *helper) sortTunings(tunings *[]v1.Tuning) {
	sort.Slice(*tunings, func(i, j int) bool {
		return (*tunings)[i].Name < (*tunings)[j].Name
	})
}

func (h *helper) genPageInfo(tunings []v1.Tuning) (v1.Page, error) {
	if !h.Page.IsRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(tunings),
			TotalItemCount: int64(len(tunings)),
		}, nil
	}

	return v1.Page{
		Total:          int64(math.Ceil(float64(len(tunings)) / float64(h.Page.Size))),
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: int64(len(tunings)),
	}, nil
}
