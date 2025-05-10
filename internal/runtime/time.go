package runtime

import (
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

func initSystemTime() error {
	time.LocalZoneSeconds = getLocalTimeZoneSeconds()
	time.LocalFixedZone = ostime.FixedZone("", time.LocalZoneSeconds)
	time.LocalZone = getLocalTimeZone()
	if time.LocalZone == "" {
		return errors.ErrInvalidTimeZone
	}

	return nil
}
