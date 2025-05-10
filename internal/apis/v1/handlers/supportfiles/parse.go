package supportfiles

import (
	"fmt"
	"strconv"
	"strings"
	ostime "time"

	query "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	duration "github.com/xhit/go-str2duration"
)

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseRoles() {
	h.roles = h.c.QueryArray("roles")
}

func (h *helper) parseHost() {
	h.host = h.c.DefaultQuery("host", "")
}

func (h *helper) parseHosts() error {
	err := h.c.ShouldBindJSON(&h.fileReq)
	if err != nil {
		return err
	}

	h.fileReq.CreatedAt = time.ISO8601Z(ostime.Now())
	return nil
}

func (h *helper) parsePast() error {
	h.past = h.c.DefaultQuery("past", "")
	if h.past == "" {
		return nil
	}

	_, err := duration.Str2Duration(h.past)
	if err != nil {
		return fmt.Errorf("invalid 'past' duration: %s", h.past)
	}

	return nil
}

func (h *helper) parsePeriod() error {
	if h.arePeriodAndPastRequired() {
		return fmt.Errorf("'past' and 'start'/'stop' cannot be used together")
	}

	qStart := h.c.DefaultQuery("start", time.RFC3339(-24*ostime.Hour))
	start, err := ostime.Parse(time.FormatRFC3339, qStart)
	if err != nil {
		return fmt.Errorf("'start' time format should be aligned with RFC3339: %s", qStart)
	}

	qStop := h.c.DefaultQuery("stop", time.NowRFC3339())
	stop, err := ostime.Parse(time.FormatRFC3339, qStop)
	if err != nil {
		return fmt.Errorf("'stop' time format should be aligned with RFC3339: %s", qStop)
	}

	h.Period = time.Period{
		Start: time.UTC(start),
		Stop:  time.UTC(stop),
	}

	return nil
}

func (h *helper) parsePage() error {
	if !query.IsPageRequired(h.c) {
		return nil
	}

	num := h.c.DefaultQuery("pageNum", "")
	if num == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	size := h.c.DefaultQuery("pageSize", "")
	if size == "" {
		return fmt.Errorf("pageSize should be provided if pageNum is provided")
	}

	var err error
	h.Page.Number, err = strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", num)
	}

	h.Page.Size, err = strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", size)
	}

	if h.Page.Number <= 0 {
		return fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if h.Page.Size <= 0 {
		return fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return nil
}

func (h *helper) arePeriodAndPastRequired() bool {
	return h.isPeriodRequired() && h.isPastRequired()
}

func (h *helper) isPeriodRequired() bool {
	return h.c.DefaultQuery("stop", "") != "" || h.c.DefaultQuery("start", "") != ""
}

func (h *helper) isPastRequired() bool {
	_, found := h.c.GetQuery("past")
	return found
}

func (h *helper) isFilterRequired() bool {
	return h.isKeywordRequired() || h.isRoleRequired() || h.isPeriodRequired()
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isRoleRequired() bool {
	return len(h.roles) > 0
}
