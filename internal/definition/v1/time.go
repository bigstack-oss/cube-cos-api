package v1

import (
	"fmt"
	"time"
)

const (
	ISO8601 = "2006-01-02T15:04:05"
	RFC3339 = time.RFC3339
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

func TimeLocal() time.Time {
	return time.Now().Local()
}

func TimeRFC3339(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(time.RFC3339)
}

func TimPastRFC3339(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(time.RFC3339)
}

func TimeLocalISO8601(t time.Time) string {
	return fmt.Sprintf("%sZ", t.Format(ISO8601))
}
