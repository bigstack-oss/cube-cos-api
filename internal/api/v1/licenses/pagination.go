package licenses

import (
	"math"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateLicenses(licenses []v1.License) []v1.License {
	if !h.page.IsRequired() {
		return licenses
	}

	left := min((h.page.Number-1)*h.page.Size, len(licenses))
	right := min(left+h.page.Size, len(licenses))
	return licenses[left:right]
}

func (h *helper) genPageInfo(licenses []v1.License) v1.Page {
	if !h.page.IsRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(licenses),
			TotalItemCount: int64(len(licenses)),
		}
	}

	return v1.Page{
		Total:          int64(math.Ceil(float64(len(licenses)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(licenses)),
	}
}
