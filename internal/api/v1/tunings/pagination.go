package tunings

import (
	"math"
	"sort"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateTunings(tunings []definition.Tuning) ([]definition.Tuning, error) {
	if !h.Page.IsRequired() {
		return tunings, nil
	}

	left := (h.Page.Number - 1) * h.Page.Size
	if left > len(tunings) {
		left = len(tunings)
	}

	right := left + h.Page.Size
	if right > len(tunings) {
		right = len(tunings)
	}

	return tunings[left:right], nil
}

func (h *helper) sortTunings(tunings *[]definition.Tuning) {
	sort.Slice(*tunings, func(i, j int) bool {
		return (*tunings)[i].Name < (*tunings)[j].Name
	})
}

func (h *helper) genPageInfo(tunings []definition.Tuning) (definition.Page, error) {
	if !h.Page.IsRequired() {
		return definition.Page{
			Total:  1,
			Number: 1,
			Size:   len(tunings),
		}, nil
	}

	return definition.Page{
		Total:  int64(math.Ceil(float64(len(tunings)) / float64(h.Page.Size))),
		Number: h.Page.Number,
		Size:   h.Page.Size,
	}, nil
}
