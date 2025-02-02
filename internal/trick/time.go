package trick

import "time"

func Minus2MinsOnMetricStartTime(start *string, t time.Time) {
	minusStart := t.Add(-2 * time.Minute).Format(time.RFC3339)
	*start = minusStart
}
