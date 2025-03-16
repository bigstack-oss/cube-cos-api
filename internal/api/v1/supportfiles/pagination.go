package supportfiles

import (
	"math"
	"sort"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateSupportFiles(supportFiles []v1.SupportFile) ([]v1.SupportFile, error) {
	if !h.Page.IsRequired() {
		return supportFiles, nil
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(supportFiles))
	right := min(left+h.Page.Size, len(supportFiles))
	return supportFiles[left:right], nil
}

func (h *helper) sortTunings(supportFiles *[]v1.SupportFile) {
	sort.Slice(*supportFiles, func(i, j int) bool {
		return (*supportFiles)[i].Name < (*supportFiles)[j].Name
	})
}

func (h *helper) genPageInfo(supportFiles []v1.SupportFile) (v1.Page, error) {
	if !h.Page.IsRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(supportFiles),
			TotalItemCount: int64(len(supportFiles)),
		}, nil
	}

	return v1.Page{
		Total:          int64(math.Ceil(float64(len(supportFiles)) / float64(h.Page.Size))),
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: int64(len(supportFiles)),
	}, nil
}
