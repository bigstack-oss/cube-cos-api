package cubecos

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/bigstack-oss/cube-cos-api/internal/prometheus"
	"github.com/stretchr/testify/require"
)

var storageTestRange = metric.Range{
	Start: ostime.Unix(1759900000, 0),
	End:   ostime.Unix(1759900060, 0),
	Step:  ostime.Minute,
}

// fakePrometheus answers each PromQL query with the value given for it, at the
// start of the test range. An unknown query fails the test.
func fakePrometheus(t *testing.T, answers map[string]string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("query")
		value, ok := answers[q]
		if !ok {
			t.Errorf("unexpected query: %s", q)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"error","error":"unexpected query"}`))
			return
		}
		_, _ = fmt.Fprintf(w, `{"status":"success","data":{"resultType":"matrix","result":[
			{"metric":{},"values":[[1759900000,%q]]}]}}`, value)
	}))
	t.Cleanup(srv.Close)
	prometheus.NewGlobalClient(srv.URL)
}

func at(v any) []metric.TimeValue {
	return []metric.TimeValue{{Time: time.LocalRFC3339(ostime.Unix(1759900000, 0)), Value: v}}
}

func TestHostsDiskBandwidthHistoryReadsOsdByteCountersFromPrometheus(t *testing.T) {
	fakePrometheus(t, map[string]string{
		"sum(rate(ceph_osd_op_r_out_bytes[4m]))": "1024.123456",
		"sum(rate(ceph_osd_op_w_in_bytes[4m]))":  "2048",
	})

	series, err := GetHostsDiskBandwidthHistory(storageTestRange)

	require.NoError(t, err)
	require.Equal(t, &metric.StorageTimeSeries{
		Unit:  "bytes",
		Read:  at(1024.1234),
		Write: at(float64(2048)),
	}, series)
}

func TestHostsDiskIopsHistoryReadsOsdOpCountersFromPrometheus(t *testing.T) {
	fakePrometheus(t, map[string]string{
		"sum(rate(ceph_osd_op_r[4m]))": "12.5",
		"sum(rate(ceph_osd_op_w[4m]))": "30",
	})

	series, err := GetHostsDiskIopsHistory(storageTestRange)

	require.NoError(t, err)
	require.Equal(t, &metric.StorageTimeSeries{
		Unit:  "ops",
		Read:  at(12.5),
		Write: at(float64(30)),
	}, series)
}

// Prometheus keeps the latency in seconds; the chart shows milliseconds.
func TestHostsDiskLatencyHistoryDividesLatencySumByCountInMilliseconds(t *testing.T) {
	fakePrometheus(t, map[string]string{
		"sum(rate(ceph_osd_op_r_latency_sum[4m])) / sum(rate(ceph_osd_op_r_latency_count[4m])) * 1000": "0.85",
		"sum(rate(ceph_osd_op_w_latency_sum[4m])) / sum(rate(ceph_osd_op_w_latency_count[4m])) * 1000": "3.2",
	})

	series, err := GetHostsDiskLatencyHistory(storageTestRange)

	require.NoError(t, err)
	require.Equal(t, &metric.StorageTimeSeries{
		Unit:  "milliseconds",
		Read:  at(0.85),
		Write: at(3.2),
	}, series)
}

// With no ops in a step the ratio is 0/0. json.Marshal rejects NaN and would
// drop the whole response, so the point must read 0.
func TestHostsDiskLatencyHistoryWithoutOpsStaysZero(t *testing.T) {
	fakePrometheus(t, map[string]string{
		"sum(rate(ceph_osd_op_r_latency_sum[4m])) / sum(rate(ceph_osd_op_r_latency_count[4m])) * 1000": "NaN",
		"sum(rate(ceph_osd_op_w_latency_sum[4m])) / sum(rate(ceph_osd_op_w_latency_count[4m])) * 1000": "NaN",
	})

	series, err := GetHostsDiskLatencyHistory(storageTestRange)

	require.NoError(t, err)
	require.Equal(t, at(float64(0)), series.Read)
	require.Equal(t, at(float64(0)), series.Write)
}

// rate() needs at least two scrapes (60s apart) inside its window, and a step
// wider than the window would skip samples between points.
func TestStorageRateWindowCoversStep(t *testing.T) {
	require.Equal(t, "4m", storageRateWindow(ostime.Minute))
	require.Equal(t, "30m", storageRateWindow(30*ostime.Minute))
	require.Equal(t, "1h", storageRateWindow(ostime.Hour))
}
