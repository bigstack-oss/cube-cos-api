package licenses

import (
	"math"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

func (h *helper) paginateLicenses(licenses []license.Options) []license.Options {
	if !h.page.IsRequired() {
		return licenses
	}

	left := min((h.page.Number-1)*h.page.Size, len(licenses))
	right := min(left+h.page.Size, len(licenses))
	return licenses[left:right]
}

func (h *helper) genPageInfo(licenses []license.Options) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(licenses),
			TotalItemCount: int64(len(licenses)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(licenses)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(licenses)),
	}
}
