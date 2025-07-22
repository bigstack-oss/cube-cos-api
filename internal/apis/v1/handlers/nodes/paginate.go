package nodes

import (
	"math"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

func (h *helper) paginateNodes(nodes []nodes.Node) []nodes.Node {
	if !h.page.IsRequired() {
		return nodes
	}

	left := min((h.page.Number-1)*h.page.Size, len(nodes))
	right := min(left+h.page.Size, len(nodes))
	return nodes[left:right]
}

func (h *helper) genPageInfo(nodes []nodes.Node) pages.Page {
	if !h.page.IsRequired() {
		return pages.Page{
			Total:          1,
			Number:         1,
			Size:           len(nodes),
			TotalItemCount: int64(len(nodes)),
		}
	}

	return pages.Page{
		Total:          int64(math.Ceil(float64(len(nodes)) / float64(h.page.Size))),
		Number:         h.page.Number,
		Size:           h.page.Size,
		TotalItemCount: int64(len(nodes)),
	}
}
