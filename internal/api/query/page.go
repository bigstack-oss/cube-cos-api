package query

import (
	"fmt"
	"strconv"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func GetPage(c *gin.Context) (*v1.Page, error) {
	num := c.DefaultQuery("pageNum", "")
	size := c.DefaultQuery("pageSize", "")
	if !IsPageRequired(num, size) {
		return nil, nil
	}

	if num == "" {
		return nil, fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if size == "" {
		return nil, fmt.Errorf("pageSize should be provided if pageNum is provided")
	}

	var err error
	page := &v1.Page{}
	page.Number, err = strconv.Atoi(num)
	if err != nil {
		return nil, fmt.Errorf("pageNum should be an integer: %s", num)
	}

	page.Size, err = strconv.Atoi(size)
	if err != nil {
		return nil, fmt.Errorf("pageSize should be an integer: %s", size)
	}

	if page.Number <= 0 {
		return nil, fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if page.Size <= 0 {
		return nil, fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return page, nil
}

func IsPageRequired(num, size string) bool {
	return num != "" || size != ""
}
