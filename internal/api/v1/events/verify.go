package events

import (
	"fmt"
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func checkQueryParams(c *gin.Context) error {
	from := c.DefaultQuery("from", definition.TimeRFC3339(-150*time.Hour))
	_, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return fmt.Errorf("invalid 'from' time: %s", from)
	}

	to := c.DefaultQuery("to", definition.TimeNowRFC3339())
	_, err = time.Parse(time.RFC3339, to)
	if err != nil {
		return fmt.Errorf("invalid 'to' time: %s", to)
	}

	return nil
}
