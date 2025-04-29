package query

import (
	"fmt"
	"time"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	duration "github.com/xhit/go-str2duration"
)

func GetPeriod(c *gin.Context) (*v1.Period, error) {
	if ArePeriodAndPastRequired(c) {
		return nil, fmt.Errorf("'past' and 'start'/'stop' cannot be used together")
	}

	timeStart := c.DefaultQuery("start", v1.TimeRFC3339(-24*time.Hour))
	start, err := time.Parse(v1.RFC3339, timeStart)
	if err != nil {
		return nil, fmt.Errorf("'start' time format should be aligned with RFC3339: %s", timeStart)
	}

	timeStop := c.DefaultQuery("stop", v1.TimeNowRFC3339())
	stop, err := time.Parse(v1.RFC3339, timeStop)
	if err != nil {
		return nil, fmt.Errorf("'stop' time format should be aligned with RFC3339: %s", timeStop)
	}

	return &v1.Period{
		Start: v1.TimeUTC(start),
		Stop:  v1.TimeUTC(stop),
	}, nil
}

func ArePeriodAndPastRequired(c *gin.Context) bool {
	return IsPeriodRequired(c) && IsPastRequired(c)
}

func IsPeriodRequired(c *gin.Context) bool {
	return c.DefaultQuery("stop", "") != "" || c.DefaultQuery("start", "") != ""
}

func IsPastRequired(c *gin.Context) bool {
	_, found := c.GetQuery("past")
	return found
}

func GetPast(c *gin.Context) (string, error) {
	query := c.DefaultQuery("past", "")
	if query == "" {
		return "", nil
	}

	_, err := duration.Str2Duration(query)
	if err != nil {
		return "", fmt.Errorf("invalid 'past' duration: %s", query)
	}

	return query, nil
}
