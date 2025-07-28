package time

import (
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

const (
	FormatBmc      = "Mon Jan 2 15:04:05 2006"
	FormatISO8601  = "2006-01-02T15:04:05"
	FormatISO8601Z = "2006-01-02T15:04:05+00:00"
	FormatRFC3339  = time.RFC3339
	FormatRFC3339Z = "2006-01-02T15:04:05Z07:00"
)

var (
	Day = 24 * time.Hour

	LocalZone        = "+00:00"
	LocalZoneSeconds = 0
	LocalFixedZone   = time.FixedZone("", LocalZoneSeconds)
)

type Period struct {
	Start string
	Stop  string
	Past  string
}

func (p *Period) InBetween(timeAt string) bool {
	t, err := time.Parse(FormatRFC3339, timeAt)
	if err != nil {
		panic(err)
	}

	start, err := time.Parse(FormatRFC3339, p.Start)
	if err != nil {
		return false
	}

	stop, err := time.Parse(FormatRFC3339, p.Stop)
	if err != nil {
		return false
	}

	return t.After(start) && t.Before(stop)
}

func NowLocal() string {
	return time.Now().Local().Format(time.RFC3339)
}

func UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func RFC3339(duration time.Duration) string {
	return time.Now().Add(duration).UTC().Format(time.RFC3339)
}

func RFC3339Z(t time.Time) string {
	return t.Format(FormatRFC3339Z)
}

func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func LocalRFC3339(t time.Time) string {
	return t.In(LocalFixedZone).Format(time.RFC3339)
}

func LocalRFC3339AddDuration(t time.Time, duration time.Duration) string {
	adjusted := t.Add(duration)
	return adjusted.In(LocalFixedZone).Format(time.RFC3339)
}

func ISO8601Z(t time.Time) string {
	return t.Format(FormatISO8601Z)
}

func Boot() string {
	bootDuration, err := host.BootTime()
	if err != nil {
		return ISO8601Z(time.Now())
	}

	bootTime := time.Unix(int64(bootDuration), 0)
	return ISO8601Z(bootTime)
}

func TimeISO8601Z(t time.Time) string {
	return t.Format(FormatISO8601Z)
}
