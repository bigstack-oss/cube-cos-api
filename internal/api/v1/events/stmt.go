package events

import (
	"fmt"
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	eventQueryTemplate = `
		from(bucket: "events")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> sort(columns: ["_time"], desc: true)
	`
)

func genQueryStmt(c *gin.Context) string {
	return fmt.Sprintf(
		eventQueryTemplate,
		c.DefaultQuery("from", definition.TimeRFC3339(-150*time.Hour)),
		c.DefaultQuery("to", definition.TimeNowRFC3339()),
		c.DefaultQuery("type", "system"),
	)
}
