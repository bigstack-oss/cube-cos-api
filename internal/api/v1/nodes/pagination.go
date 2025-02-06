package nodes

import (
	"math"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func paginateNodes(nodes []*definition.Node, page definition.Page) ([]*definition.Node, error) {
	if !page.IsRequired() {
		return nodes, nil
	}

	left := (page.Number - 1) * page.Size
	if left > len(nodes) {
		left = len(nodes)
	}

	right := left + page.Size
	if right > len(nodes) {
		right = len(nodes)
	}

	return nodes[left:right], nil
}

func genPageInfo(nodes []*definition.Node, page definition.Page) (definition.Page, error) {
	if !page.IsRequired() {
		return definition.Page{
			Total:  1,
			Number: 1,
			Size:   len(nodes),
		}, nil
	}

	return definition.Page{
		Total:  int64(math.Ceil(float64(len(nodes)) / float64(page.Size))),
		Number: page.Number,
		Size:   page.Size,
	}, nil
}
