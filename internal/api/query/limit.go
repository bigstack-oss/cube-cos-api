package query

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetLimit(c *gin.Context) (int, error) {
	query := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(query)
	if err != nil {
		return 0, err
	}

	if limit <= 0 {
		return 0, fmt.Errorf("limit should be greater than 0")
	}

	return limit, nil
}
