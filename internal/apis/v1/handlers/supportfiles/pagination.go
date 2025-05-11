package supportfiles

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

func (h *helper) paginateFileSets(fileSets []support.FileSet) ([]support.FileSet, error) {
	if !h.Page.IsRequired() {
		return fileSets, nil
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(fileSets))
	right := min(left+h.Page.Size, len(fileSets))
	return fileSets[left:right], nil
}

func (h *helper) sortFileSets(fileSets *[]support.FileSet) {
	sort.Slice(*fileSets, func(i, j int) bool {
		return (*fileSets)[i].Status.CreatedAt < (*fileSets)[j].Status.CreatedAt
	})
}

func (h *helper) genPageInfo(fileSets []support.FileSet) (pages.Page, error) {
	if !h.Page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(fileSets),
			TotalItemCount: int64(len(fileSets)),
		}, nil
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(fileSets)) / float64(h.Page.Size))),
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: int64(len(fileSets)),
	}, nil
}
