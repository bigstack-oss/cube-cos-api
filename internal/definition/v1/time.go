package v1

import "time"

const (
	Iso8601 = "2006-01-02T15:04:05"
	RFC3339 = time.RFC3339
)

func TimeNowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func TimeRFC3339(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(time.RFC3339)
}
