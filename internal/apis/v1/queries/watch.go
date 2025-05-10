package queries

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetWatch(c *gin.Context) (bool, error) {
	query := c.DefaultQuery("watch", "false")
	watch, err := strconv.ParseBool(query)
	if err != nil {
		return false, errors.New("watch parameter is invalid, it should be true or false if provided")
	}

	return watch, nil
}
