package licenses

import (
	"math"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateLicenses(licenses []v1.License) []v1.License {
	if !h.Page.IsRequired() {
		return licenses
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(licenses))
	right := min(left+h.Page.Size, len(licenses))
	return licenses[left:right]
}

func (h *helper) genPageInfo(licenses []v1.License) v1.Page {
	if !h.Page.IsRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(licenses),
			TotalItemCount: int64(len(licenses)),
		}
	}

	return v1.Page{
		Total:          int64(math.Ceil(float64(len(licenses)) / float64(h.Page.Size))),
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: int64(len(licenses)),
	}
}
