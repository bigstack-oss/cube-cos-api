package nodes

import (
	"math"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func paginateNodes(nodes []*definition.Node, page definition.Page) ([]*definition.Node, error) {
	if !page.IsRequired() {
		return nodes, nil
	}

	left := min((page.Number-1)*page.Size, len(nodes))
	right := min(left+page.Size, len(nodes))
	return nodes[left:right], nil
}

func genPageInfo(nodes []*definition.Node, page definition.Page) (definition.Page, error) {
	if !page.IsRequired() {
		return definition.Page{
			Total:          1,
			Number:         1,
			Size:           len(nodes),
			TotalItemCount: int64(len(nodes)),
		}, nil
	}

	return definition.Page{
		Total:          int64(math.Ceil(float64(len(nodes)) / float64(page.Size))),
		Number:         page.Number,
		Size:           page.Size,
		TotalItemCount: int64(len(nodes)),
	}, nil
}
