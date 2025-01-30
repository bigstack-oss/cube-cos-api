package summary

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseWatch(c *gin.Context) (bool, error) {
	rawParam := c.DefaultQuery("watch", "false")
	watch, err := strconv.ParseBool(rawParam)
	if err != nil {
		return false, errors.New("watch parameter is invalid, it should be true or false if provided")
	}

	return watch, nil
}
