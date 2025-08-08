package queries

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	ostime "time"

	bserrors "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	duration "github.com/xhit/go-str2duration"
)

func GetReqId(c *gin.Context) string {
	id, found := c.Get("reqId")
	if !found {
		return ""
	}

	return id.(string)
}

func GetKeyword(c *gin.Context) string {
	keyword := c.DefaultQuery("keyword", "")
	return strings.ToLower(keyword)
}

func GetLimit(c *gin.Context, defaultLimit int) (int, error) {
	query := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	limit, err := strconv.Atoi(query)
	if err != nil {
		return 0, err
	}

	if limit <= 0 {
		return 0, bserrors.ErrLimitInvalid
	}

	return limit, nil
}

func GetPage(c *gin.Context) (*pages.Page, error) {
	if !IsPageRequired(c) {
		return &pages.Page{}, nil
	}

	num := c.DefaultQuery("pageNum", "")
	if num == "" {
		return nil, fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	size := c.DefaultQuery("pageSize", "")
	if size == "" {
		return nil, fmt.Errorf("pageSize should be provided if pageNum is provided")
	}

	var err error
	page := &pages.Page{}
	page.Number, err = strconv.Atoi(num)
	if err != nil {
		return nil, fmt.Errorf("pageNum should be an integer: %s", num)
	}

	page.Size, err = strconv.Atoi(size)
	if err != nil {
		return nil, fmt.Errorf("pageSize should be an integer: %s", size)
	}

	if page.Number <= 0 {
		return nil, fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if page.Size <= 0 {
		return nil, fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return page, nil
}

func IsPageRequired(c *gin.Context) bool {
	return c.DefaultQuery("pageNum", "") != "" || c.DefaultQuery("pageSize", "") != ""
}

func ParseRecordRequire(c *gin.Context) bool {
	val, found := c.GetQuery("isRecordRequired")
	if !found {
		return true
	}

	return val == "true"
}

func GetPeriod(c *gin.Context) (*time.Period, error) {
	if ArePeriodAndPastRequired(c) {
		return nil, fmt.Errorf("'past' and 'start'/'stop' cannot be used together")
	}

	defaultStart := time.LocalRFC3339AddDuration(ostime.Now().Local(), -24*ostime.Hour)
	timeStart := c.DefaultQuery("start", defaultStart)
	_, err := ostime.Parse(time.FormatRFC3339, timeStart)
	if err != nil {
		return nil, fmt.Errorf("'start' time format should be aligned with RFC3339: %s", timeStart)
	}

	defaultStop := time.LocalRFC3339(ostime.Now())
	timeStop := c.DefaultQuery("stop", defaultStop)
	_, err = ostime.Parse(time.FormatRFC3339, timeStop)
	if err != nil {
		return nil, fmt.Errorf("'stop' time format should be aligned with RFC3339: %s", timeStop)
	}

	return &time.Period{
		Start: timeStart,
		Stop:  timeStop,
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

func GetPastTime(c *gin.Context) (string, error) {
	past, err := GetPast(c)
	if err != nil {
		return "", err
	}

	if past == "" {
		past = "1h"
	}

	duration, err := duration.Str2Duration(past)
	if err != nil {
		return "", fmt.Errorf("invalid 'past' duration: %s", past)
	}

	t := ostime.Now().Add(-duration)
	return time.LocalRFC3339(t), nil
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
