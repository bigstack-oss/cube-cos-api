package notifications

import (
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	log "go-micro.dev/v5/logger"
)

func (h *helper) convertListOpts() (*notifications.ListOpts, error) {
	start, err := ostime.Parse(time.FormatRFC3339, h.period.Start)
	if err != nil {
		log.Errorf("notifications(%s): failed to convert start time(%v)", h.reqId, err)
		return nil, err
	}

	stop, err := ostime.Parse(time.FormatRFC3339, h.period.Stop)
	if err != nil {
		log.Errorf("notifications(%s): failed to convert stop time(%v)", h.reqId, err)
		return nil, err
	}

	past, err := ostime.Parse(time.FormatRFC3339, h.past)
	if err != nil {
		log.Errorf("notifications(%s): failed to convert past time(%v)", h.reqId, err)
		return nil, err
	}

	opts := &notifications.ListOpts{
		Limit:     int64(h.limit),
		Desending: true,
		Start:     past,
		Stop:      ostime.Now(),
	}

	if queries.IsPeriodRequired(h.c) {
		opts.Start = start
		opts.Stop = stop
	}

	return opts, nil
}
