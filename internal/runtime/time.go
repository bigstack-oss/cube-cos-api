package runtime

import (
	"time"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
)

func initSystemTime() error {
	v1.LocalTimeZoneSeconds = getLocalTimeZoneSeconds()
	v1.LocalTimeFixedZone = time.FixedZone("", v1.LocalTimeZoneSeconds)
	v1.LocalTimeZone = getLocalTimeZone()
	if v1.LocalTimeZone == "" {
		return errors.ErrInvalidTimeZone
	}

	return nil
}
