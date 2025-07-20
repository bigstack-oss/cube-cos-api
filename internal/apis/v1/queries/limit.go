package queries

import (
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/gin-gonic/gin"
)

func GetLimit(c *gin.Context, defaultLimit int) (int, error) {
	query := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	limit, err := strconv.Atoi(query)
	if err != nil {
		return 0, err
	}

	if limit <= 0 {
		return 0, errors.ErrLimitInvalid
	}

	return limit, nil
}
