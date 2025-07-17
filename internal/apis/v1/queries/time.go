package queries

import (
	"errors"
	"fmt"
	"strconv"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	duration "github.com/xhit/go-str2duration"
)

func GetPeriod(c *gin.Context) (*time.Period, error) {
	if ArePeriodAndPastRequired(c) {
		return nil, fmt.Errorf("'past' and 'start'/'stop' cannot be used together")
	}

	timeStart := c.DefaultQuery("start", time.RFC3339(-24*ostime.Hour))
	start, err := ostime.Parse(time.FormatRFC3339, timeStart)
	if err != nil {
		return nil, fmt.Errorf("'start' time format should be aligned with RFC3339: %s", timeStart)
	}

	timeStop := c.DefaultQuery("stop", time.NowRFC3339())
	stop, err := ostime.Parse(time.FormatRFC3339, timeStop)
	if err != nil {
		return nil, fmt.Errorf("'stop' time format should be aligned with RFC3339: %s", timeStop)
	}

	return &time.Period{
		Start: time.UTC(start),
		Stop:  time.UTC(stop),
	}, nil
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

func GetAggregation(c *gin.Context) (bool, error) {
	query := c.DefaultQuery("aggregate", "false")
	aggregation, err := strconv.ParseBool(query)
	if err != nil {
		return false, errors.New("aggregate parameter is invalid, it should be true or false if provided")
	}

	return aggregation, nil
}

func ArePeriodAndPastRequired(c *gin.Context) bool {
	return IsPeriodRequired(c) && IsPastRequired(c)
}

func ArePeriodAndPastEmpty(c *gin.Context) bool {
	return !IsPeriodRequired(c) && !IsPastRequired(c)
}

func IsPeriodRequired(c *gin.Context) bool {
	return c.DefaultQuery("stop", "") != "" || c.DefaultQuery("start", "") != ""
}

func IsPastRequired(c *gin.Context) bool {
	_, found := c.GetQuery("past")
	return found
}

func IsAggregationRequired(c *gin.Context) bool {
	_, found := c.GetQuery("aggregation")
	return found
}
