package events

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
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

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch data center list successfully",
		"data":   events,
	})
}

func genQueryStmt(c *gin.Context) string {
	measurement := c.DefaultQuery("type", "system")
	from := c.DefaultQuery("from", time.Now().Add(-1*time.Hour).UTC().String())
	to := c.DefaultQuery("to", time.Now().UTC().String())

	return fmt.Sprintf(
		eventQueryTemplate,
		from,
		to,
		measurement,
	)
}
