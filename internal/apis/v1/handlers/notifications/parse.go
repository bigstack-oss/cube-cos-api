package notifications

import "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"

func (h *helper) parseParamByHandler() error {
	switch h.handler {
	case "listNotifications":
		return h.parseListNotificationParams()
	default:
		return nil
	}
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) parseListNotificationParams() error {
	var err error
	h.limit, err = queries.GetLimit(h.c, 100)
	if err != nil {
		return err
	}

	h.past, err = queries.GetPastTime(h.c)
	if err != nil {
		return err
	}

	h.period, err = queries.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.page, err = queries.GetPage(h.c)
	if err != nil {
		return err
	}

	h.keyword = queries.GetKeyword(h.c)
	return nil
}
