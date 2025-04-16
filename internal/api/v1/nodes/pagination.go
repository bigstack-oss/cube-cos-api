package nodes

import (
	"math"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) paginateNodes(nodes []definition.Node) []definition.Node {
	if !h.Page.IsRequired() {
		return nodes
	}

	left := min((h.Page.Number-1)*h.Page.Size, len(nodes))
	right := min(left+h.Page.Size, len(nodes))
	return nodes[left:right]
}

func genPageInfo(nodes []definition.Node, page definition.Page) definition.Page {
	if !page.IsRequired() {
		return definition.Page{
			Total:          1,
			Number:         1,
			Size:           len(nodes),
			TotalItemCount: int64(len(nodes)),
		}
	}

	return definition.Page{
		Total:          int64(math.Ceil(float64(len(nodes)) / float64(page.Size))),
		Number:         page.Number,
		Size:           page.Size,
		TotalItemCount: int64(len(nodes)),
	}
}
