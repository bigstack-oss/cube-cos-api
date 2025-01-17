package v1

import "time"

const (
	Iso8601 = "2006-01-02T15:04:05"
)

func TimeNowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
