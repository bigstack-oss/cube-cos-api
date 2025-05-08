package nodes

import (
	"math"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateNodes(nodes []v1.Node) []v1.Node {
	if !h.Page.IsRequired() {
		return nodes
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(nodes))
	right := min(left+h.Page.Size, len(nodes))
	return nodes[left:right]
}

func (h *helper) genPageInfo(nodes []v1.Node) v1.Page {
	if !h.Page.IsRequired() {
		return v1.Page{
			Total:          1,
			Number:         1,
			Size:           len(nodes),
			TotalItemCount: int64(len(nodes)),
		}
	}

	return v1.Page{
		Total:          int64(math.Ceil(float64(len(nodes)) / float64(h.Page.Size))),
		Number:         h.Page.Number,
		Size:           h.Page.Size,
		TotalItemCount: int64(len(nodes)),
	}
}
