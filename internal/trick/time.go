package trick

import "time"

func Minus2MinsOnMetricStart(t time.Time) time.Time {
	return t.Add(-2 * time.Minute)
}
