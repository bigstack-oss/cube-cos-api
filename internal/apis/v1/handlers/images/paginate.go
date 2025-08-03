package images

import (
	"math"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type imagePage struct {
	Images     []images.Image `json:"images"`
	pages.Page `json:"page"`
}

func (h *helper) paginateImages(images []images.Image) []images.Image {
	if !h.page.IsRequired() {
		return images
	}

	left := min((h.page.Number-1)*h.page.Size, len(images))
	right := min(left+h.page.Size, len(images))
	return images[left:right]
}

func (h *helper) sortimages(images *[]images.Image) {
	sort.Slice(*images, func(i, j int) bool {
		return (*images)[i].CreatedAt > (*images)[j].CreatedAt
	})
}

func (h *helper) genPageInfo(images []images.Image) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(images),
			TotalItemCount: int64(len(images)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(images)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(images)),
	}
}
