package events

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events",
			Func:    getEvents,
		},
	}
)

func getEvents(c *gin.Context) {
	stmt := genQueryStmt(c)
	events, err := cubecos.ListEvents(stmt)
	if err != nil {
		logger.Errorf("failed to fetch events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":   http.StatusInternalServerError,
			"status": "internal server error",
			"msg":    "failed to fetch events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch events successfully",
		"data":   events,
	})
}

func genQueryStmt(c *gin.Context) string {
	measurement := c.DefaultQuery("type", "system")
	from := c.DefaultQuery("from", definition.TimeRFC3339(-72*time.Hour))
	to := c.DefaultQuery("to", definition.TimeNowRFC3339())

	return fmt.Sprintf(
		eventQueryTemplate,
		from,
		to,
		measurement,
	)
}
