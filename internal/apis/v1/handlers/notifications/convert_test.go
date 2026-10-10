package notifications

import (
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	ostime "time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newListContext(t *testing.T, query url.Values) *gin.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(nethttp.MethodGet, "/notifications?"+query.Encode(), nil)

	return c
}

func newListHelper(t *testing.T, query url.Values) *helper {
	t.Helper()

	h, err := initHepler(newListContext(t, query), "listNotifications")
	require.NoError(t, err)

	return h
}

func TestConvertListOptsKeepsTheInstantsOfStartAndStop(t *testing.T) {
	wantStart := ostime.Date(2026, 10, 9, 16, 3, 42, 0, ostime.UTC)
	wantStop := ostime.Date(2026, 10, 9, 17, 0, 0, 0, ostime.UTC)

	tests := []struct {
		name  string
		start string
		stop  string
	}{
		{name: "Z", start: "2026-10-09T16:03:42Z", stop: "2026-10-09T17:00:00Z"},
		{name: "+00:00", start: "2026-10-09T16:03:42+00:00", stop: "2026-10-09T17:00:00+00:00"},
		{name: "+08:00", start: "2026-10-10T00:03:42+08:00", stop: "2026-10-10T01:00:00+08:00"},
		{name: "+09:00", start: "2026-10-10T01:03:42+09:00", stop: "2026-10-10T02:00:00+09:00"},
		{name: "-05:00", start: "2026-10-09T11:03:42-05:00", stop: "2026-10-09T12:00:00-05:00"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newListHelper(t, url.Values{"start": {test.start}, "stop": {test.stop}})

			opts, err := h.convertListOpts()
			require.NoError(t, err)
			require.Truef(t, opts.Start.Equal(wantStart), "start %s", opts.Start)
			require.Truef(t, opts.Stop.Equal(wantStop), "stop %s", opts.Stop)
			require.True(t, opts.Desending)
		})
	}
}

func TestConvertListOptsDefaultsToThePastHourUpToNow(t *testing.T) {
	h := newListHelper(t, url.Values{})

	opts, err := h.convertListOpts()
	require.NoError(t, err)
	require.WithinDuration(t, ostime.Now().Add(-ostime.Hour), opts.Start, 5*ostime.Second)
	require.WithinDuration(t, ostime.Now(), opts.Stop, 5*ostime.Second)
	require.Equal(t, int64(100), opts.Limit)
}
