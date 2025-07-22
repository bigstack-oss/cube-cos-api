package tunings

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
)

func (h *helper) paginateTunings(tunings []tunings.Tuning) ([]tunings.Tuning, error) {
	if !h.Page.IsRequired() {
		return tunings, nil
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(tunings))
	right := min(left+h.Page.Size, len(tunings))
	return tunings[left:right], nil
}

func (h *helper) sortTunings(tunings *[]tunings.Tuning) {
	for i := range *tunings {
		h.sortHosts(&(*tunings)[i])
	}

	sort.Slice(*tunings, func(i, j int) bool {
		if (*tunings)[i].Name == (*tunings)[j].Name {
			return len((*tunings)[i].Hosts) < len((*tunings)[j].Hosts)
		}

		return (*tunings)[i].Name < (*tunings)[j].Name
	})
}

func (h *helper) sortHosts(tuning *tunings.Tuning) {
	sort.Slice(tuning.Hosts, func(i, j int) bool {
		return tuning.Hosts[i].Name < tuning.Hosts[j].Name
	})
}

func (h *helper) genPageInfo(tunings []tunings.Tuning) (pages.Page, error) {
	if !h.Page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(tunings),
			TotalItemCount: int64(len(tunings)),
		}, nil
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(tunings)) / float64(h.Page.Size))),
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: int64(len(tunings)),
	}, nil
}
