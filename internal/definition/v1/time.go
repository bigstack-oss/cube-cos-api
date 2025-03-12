package v1

import (
	"fmt"
	"time"
)

const (
	ISO8601  = "2006-01-02T15:04:05"
	ISO8601Z = "2006-01-02T15:04:05+00:00"
	RFC3339  = time.RFC3339
	RFC3339Z = "2006-01-02T15:04:05Z07:00"
)

var (
	LocalTimeZone        = "+00:00"
	LocalTimeZoneSeconds = 0
	LocalTimeFixedZone   = time.FixedZone("", LocalTimeZoneSeconds)
)

type Period struct {
	Start string
	Stop  string
}

func TimeUTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func TimeNowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func TimeNowRFC3339Z() string {
	return time.Now().UTC().Format(RFC3339Z)
}

func TimeLocal() string {
	return time.Now().Local().Format(time.RFC3339)
}

func TimeRFC3339(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(time.RFC3339)
}

func TimeRFC3339Z(t time.Time) string {
	return t.Format(RFC3339Z)
}

func TimeLocalRFC3339(t time.Time) string {
	return t.In(LocalTimeFixedZone).Format(time.RFC3339)
}

func TimPastRFC3339(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(time.RFC3339)
}

func TimePastRFC3339Z(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(RFC3339Z)
}

func TimeLocalISO8601(t time.Time) string {
	return fmt.Sprintf("%sZ", t.Format(ISO8601))
}

func TimeISO8601Z(t time.Time) string {
	return t.Format(ISO8601Z)
}
