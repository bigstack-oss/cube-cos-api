package notifications

import "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"

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

	return nil
}
