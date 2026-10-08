package runtime

import (
	"testing"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/stretchr/testify/require"
)

// With no url in the config, the API reaches Prometheus the way it reaches
// InfluxDB: through the data center VIP, on haproxy's /prometheus route.
func TestPrometheusUrlDefaultsToDataCenterVip(t *testing.T) {
	t.Cleanup(func() { conf.Opts.Spec.Store.Prometheus.Url = "" })
	vip := base.DataCenterVip
	t.Cleanup(func() { base.DataCenterVip = vip })
	base.DataCenterVip = "10.32.36.10"

	require.Equal(t, "http://10.32.36.10/prometheus", parsePrometheusUrl())
}

func TestPrometheusUrlFromConfigWins(t *testing.T) {
	t.Cleanup(func() { conf.Opts.Spec.Store.Prometheus.Url = "" })
	conf.Opts.Spec.Store.Prometheus.Url = "http://127.0.0.1:9091/prometheus"

	require.Equal(t, "http://127.0.0.1:9091/prometheus", parsePrometheusUrl())
}
