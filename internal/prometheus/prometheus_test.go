package prometheus

import (
	osmath "math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newServer(t *testing.T, body string, check func(r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestQueryRangeSendsRangeAndParsesMatrix(t *testing.T) {
	start := time.Unix(1759900000, 0)
	end := start.Add(2 * time.Minute)

	srv := newServer(t, `{"status":"success","data":{"resultType":"matrix","result":[
		{"metric":{},"values":[[1759900000,"10.5"],[1759900060,"20"]]}]}}`,
		func(r *http.Request) {
			require.Equal(t, "/prometheus/api/v1/query_range", r.URL.Path)
			require.Equal(t, "sum(rate(ceph_osd_op_r[4m]))", r.URL.Query().Get("query"))
			require.Equal(t, "1759900000", r.URL.Query().Get("start"))
			require.Equal(t, "1759900120", r.URL.Query().Get("end"))
			require.Equal(t, "60", r.URL.Query().Get("step"))
		})

	c := NewClient(srv.URL + "/prometheus")
	samples, err := c.QueryRange("sum(rate(ceph_osd_op_r[4m]))", start, end, time.Minute)

	require.NoError(t, err)
	require.Equal(t, []Sample{
		{Time: time.Unix(1759900000, 0), Value: 10.5},
		{Time: time.Unix(1759900060, 0), Value: 20},
	}, samples)
}

// No OSD series in the window (scrape down, fresh cluster) gives an empty
// matrix. That is an empty chart, not an error.
func TestQueryRangeWithEmptyMatrixGivesNoSamples(t *testing.T) {
	srv := newServer(t, `{"status":"success","data":{"resultType":"matrix","result":[]}}`, nil)

	samples, err := NewClient(srv.URL).QueryRange("up", time.Unix(0, 0), time.Unix(60, 0), time.Minute)

	require.NoError(t, err)
	require.Empty(t, samples)
}

// A latency ratio with no ops in a step comes back as the string "NaN".
func TestQueryRangeKeepsNaNSamples(t *testing.T) {
	srv := newServer(t, `{"status":"success","data":{"resultType":"matrix","result":[
		{"metric":{},"values":[[60,"NaN"]]}]}}`, nil)

	samples, err := NewClient(srv.URL).QueryRange("x", time.Unix(0, 0), time.Unix(60, 0), time.Minute)

	require.NoError(t, err)
	require.Len(t, samples, 1)
	require.True(t, osmath.IsNaN(samples[0].Value))
}

func TestQueryRangeReturnsPrometheusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":"error","errorType":"bad_data","error":"parse error"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := NewClient(srv.URL).QueryRange("sum(", time.Unix(0, 0), time.Unix(60, 0), time.Minute)

	require.ErrorContains(t, err, "parse error")
}
