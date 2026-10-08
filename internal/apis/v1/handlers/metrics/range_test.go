package metrics

import (
	"net/http/httptest"
	"testing"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestHelper(t *testing.T, rawQuery string) *helper {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/metrics/diskIops/history/hosts?"+rawQuery, nil)
	return &helper{c: c}
}

// The home page sends start/stop. The range must follow them, so the chart
// shows the hour the page asked for.
func TestStorageRangeFollowsStartAndStop(t *testing.T) {
	h := newTestHelper(t, "start=2026-10-08T09:00:00%2B08:00&stop=2026-10-08T10:00:00%2B08:00")
	h.Period = &time.Period{Start: "2026-10-08T09:00:00+08:00", Stop: "2026-10-08T10:00:00+08:00"}
	h.aggregateWindow = "1m"

	r, err := h.genStorageRange()

	require.NoError(t, err)
	require.True(t, r.Start.Equal(ostime.Date(2026, 10, 8, 1, 0, 0, 0, ostime.UTC)))
	require.True(t, r.End.Equal(ostime.Date(2026, 10, 8, 2, 0, 0, 0, ostime.UTC)))
	require.Equal(t, ostime.Minute, r.Step)
}

func TestStorageRangeFollowsPast(t *testing.T) {
	h := newTestHelper(t, "past=2h")
	h.past = "2h"
	h.aggregateWindow = "1m"

	before := ostime.Now()
	r, err := h.genStorageRange()
	after := ostime.Now()

	require.NoError(t, err)
	require.False(t, r.End.Before(before.Truncate(ostime.Second)))
	require.False(t, r.End.After(after))
	require.Equal(t, 2*ostime.Hour, r.End.Sub(r.Start))
	require.Equal(t, ostime.Minute, r.Step)
}

func TestStorageRangeUsesAggregateWindowAsStep(t *testing.T) {
	h := newTestHelper(t, "past=48h")
	h.past = "48h"
	h.aggregateWindow = "1h"

	r, err := h.genStorageRange()

	require.NoError(t, err)
	require.Equal(t, ostime.Hour, r.Step)
}
