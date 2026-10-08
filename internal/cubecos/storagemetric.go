package cubecos

import (
	"fmt"
	osmath "math"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/bigstack-oss/cube-cos-api/internal/prometheus"
	log "go-micro.dev/v5/logger"
)

// The storage charts read the per-OSD counters that ceph-mgr's prometheus module
// exports. They used to read the "ceph" InfluxDB bucket, whose writer (the mgr
// influx module) was retired, and the PromQL here matches the Grafana storage
// dashboard that moved off it.

// Prometheus scrapes ceph every 60s. rate() needs two samples in its window, and
// four scrapes keep a missed one from blanking a point.
const minStorageRateWindow = 4 * ostime.Minute

func GetHostsDiskBandwidthHistory(r metric.Range) (*metric.StorageTimeSeries, error) {
	return getStorageTimeSeries(r, "bytes",
		"sum(rate(ceph_osd_op_r_out_bytes[%[1]s]))",
		"sum(rate(ceph_osd_op_w_in_bytes[%[1]s]))",
	)
}

func GetHostsDiskIopsHistory(r metric.Range) (*metric.StorageTimeSeries, error) {
	return getStorageTimeSeries(r, "ops",
		"sum(rate(ceph_osd_op_r[%[1]s]))",
		"sum(rate(ceph_osd_op_w[%[1]s]))",
	)
}

// GetHostsDiskLatencyHistory gives the mean op latency across all OSDs. The
// counters are in seconds; the chart shows milliseconds.
func GetHostsDiskLatencyHistory(r metric.Range) (*metric.StorageTimeSeries, error) {
	return getStorageTimeSeries(r, "milliseconds",
		"sum(rate(ceph_osd_op_r_latency_sum[%[1]s])) / sum(rate(ceph_osd_op_r_latency_count[%[1]s])) * 1000",
		"sum(rate(ceph_osd_op_w_latency_sum[%[1]s])) / sum(rate(ceph_osd_op_w_latency_count[%[1]s])) * 1000",
	)
}

func getStorageTimeSeries(r metric.Range, unit, readQuery, writeQuery string) (*metric.StorageTimeSeries, error) {
	window := storageRateWindow(r.Step)

	read, err := getStorageHistory(r, fmt.Sprintf(readQuery, window))
	if err != nil {
		log.Errorf("metrics: failed to get storage read %s series(%v)", unit, err)
		return nil, err
	}

	write, err := getStorageHistory(r, fmt.Sprintf(writeQuery, window))
	if err != nil {
		log.Errorf("metrics: failed to get storage write %s series(%v)", unit, err)
		return nil, err
	}

	return &metric.StorageTimeSeries{
		Unit:  unit,
		Read:  read,
		Write: write,
	}, nil
}

func getStorageHistory(r metric.Range, query string) ([]metric.TimeValue, error) {
	samples, err := prometheus.QueryRange(query, r.Start, r.End, r.Step)
	if err != nil {
		return nil, err
	}

	points := make([]metric.TimeValue, 0, len(samples))
	for _, s := range samples {
		// 0/0 when no op ran in the step. json.Marshal rejects NaN.
		value := s.Value
		if osmath.IsNaN(value) || osmath.IsInf(value, 0) {
			value = 0
		}

		points = append(points, metric.TimeValue{
			Time:  time.LocalRFC3339(s.Time),
			Value: math.RoundDown(value, 4),
		})
	}

	return points, nil
}

func storageRateWindow(step ostime.Duration) string {
	window := max(step, minStorageRateWindow)
	switch {
	case window%ostime.Hour == 0:
		return fmt.Sprintf("%dh", window/ostime.Hour)
	case window%ostime.Minute == 0:
		return fmt.Sprintf("%dm", window/ostime.Minute)
	default:
		return fmt.Sprintf("%ds", window/ostime.Second)
	}
}
