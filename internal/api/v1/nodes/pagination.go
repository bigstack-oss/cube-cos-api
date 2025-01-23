package nodes

import (
	"fmt"
	"math"
	"strconv"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func genPageOptsByQueryParams(c *gin.Context) (definition.Page, error) {
	var err error
	page := definition.Page{}
	pageNum := c.DefaultQuery("pageNum", "")
	pageSize := c.DefaultQuery("pageSize", "")

	if !definition.IsPaginationRequired(pageNum, pageSize) {
		return page, nil
	}

	if pageNum == "" {
		return page, fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if pageSize == "" {
		return page, fmt.Errorf("pageSize should greater than 0 if pageNum is provided")
	}

	page.Number, err = strconv.Atoi(pageNum)
	if err != nil {
		return page, fmt.Errorf("pageNum should be an integer: %s", pageNum)
	}

	page.Size, err = strconv.Atoi(pageSize)
	if err != nil {
		return page, fmt.Errorf("pageSize should be an integer: %s", pageSize)
	}

	if page.Number <= 0 {
		return page, fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if page.Size <= 0 {
		return page, fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return page, nil
}

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
