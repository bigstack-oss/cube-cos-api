package metrics

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
)

func (h *helper) genHostsCpuSummaryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(`fn: (r) => r._measurement == "cpu" and r._field == "usage_idle"`).
		AggregateWindow(`every: 60s, fn: mean, createEmpty: false`).
		Map(`fn: (r) => ({ r with _value: 100.0 - r._value })`).
		Last().
		String()
}

// note:
// will be supported in the M2
func (h *helper) genHostMemoryUsageStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(`fn: (r) => r._measurement == "mem" and (r._field == "used" or r._field == "total"`).
		AggregateWindow(`every: 60s, fn: mean, createEmpty: false`).
		Pivot(`rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value"`).
		Map(`fn: (r) => ({ r with _value: (r.used * 100.0) / r.total })`).
		Last().
		String()
}

func (h *helper) genHostMemoryUsageRankStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(`fn: (r) => r._measurement == "mem" and r._field == "used_percent" and r.role == "cube"`).
		Last().
		Group(`columns: ["host"]`).
		Top(`n: 10, columns: ["_value"]`).
		String()
}

func (h *helper) genHostCpuUsageRankStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(`fn: (r) => r._measurement == "cpu" and r._field == "usage_idle" and r.role == "cube"`).
		Group(`columns: ["host"]`).
		Last().
		Map(`fn: (r) => ({ r with used: 100.0 - r._value })`).
		Group("").
		Top(fmt.Sprintf(`n: %d, columns: ["used"]`, h.rank.head)).
		Keep(`columns: ["host", "used"]`).
		String()
}

func (h *helper) genHostsDiskReadBandwidthStmt() string {
	query := influx.Query{}
	return query.Bucket("ceph").
		Range(h.genTimeDuration()).
		Filter(`fn: (r) => r._measurement == "ceph_daemon_stats" and r.ceph_daemon =~ /^osd\.[0-9]+$/ and r.type_instance == "osd.op_r_out_bytes"`).
		AggregateWindow(`every: 60s, fn: sum, createEmpty: false`).
		Derivative(`unit: 1s, nonNegative: true`).
		Group(`columns: ["_time"]`).
		Max(`column: "_value"`).
		Group("").
		String()
}

func (h *helper) genHostsDiskWriteBandwidthStmt() string {
	query := influx.Query{}
	return query.Bucket("ceph").
		Range(h.genTimeDuration()).
		Filter(`fn: (r) => r._measurement == "ceph_daemon_stats" and r.ceph_daemon =~ /^osd\.[0-9]+$/ and r.type_instance == "osd.op_w_in_bytes"`).
		AggregateWindow(`every: 60s, fn: sum, createEmpty: false`).
		Derivative(`unit: 1s, nonNegative: true`).
		Group(`columns: ["_time"]`).
		Max(`column: "_value"`).
		Group("").
		String()
}

func (h *helper) genHostStorageReadIopsStmt() string {
	query := influx.Query{}
	return query.Bucket("ceph").
		Range(h.genTimeDuration()).
		Filter(`fn: (r) => r._measurement == "ceph_daemon_stats" and r.ceph_daemon =~ /^osd\.[0-9]+$/ and r.type_instance == "osd.op_r"`).
		AggregateWindow(`every: 60s, fn: sum, createEmpty: false`).
		Derivative(`unit: 1s, nonNegative: true`).
		Group(`columns: ["_time"]`).
		Max(`column: "_value"`).
		Group("").
		String()
}

func (h *helper) genHostStorageWriteIopsStmt() string {
	query := influx.Query{}
	return query.Bucket("ceph").
		Range(h.genTimeDuration()).
		Filter(`fn: (r) => r._measurement == "ceph_daemon_stats" and r.ceph_daemon =~ /^osd\.[0-9]+$/ and r.type_instance == "osd.op_w"`).
		AggregateWindow(`every: 60s, fn: sum, createEmpty: false`).
		Derivative(`unit: 1s, nonNegative: true`).
		Group(`columns: ["_time"]`).
		Max(`column: "_value"`).
		Group("").
		String()
}

func (h *helper) genHostStorageReadLatencyStmt() string {
	query := influx.Query{}
	return query.Bucket("ceph").
		Range(h.genTimeDuration()).
		Filter(`fn: (r) => r._measurement == "ceph_daemon_stats" and r.ceph_daemon =~ /^osd\.[0-9]+$/ and r.type_instance == "osd.op_r_latency"`).
		AggregateWindow(`every: 60s, fn: sum, createEmpty: false`).
		Different().
		Derivative(`unit: 1s, nonNegative: true`).
		Group(`columns: ["_time"]`).
		Max(`column: "_value"`).
		Group("").
		String()
}

func (h *helper) genHostStorageWriteLatencyStmt() string {
	query := influx.Query{}
	return query.Bucket("ceph").
		Range(h.genTimeDuration()).
		Filter(`fn: (r) => r._measurement == "ceph_daemon_stats" and r.ceph_daemon =~ /^osd\.[0-9]+$/ and r.type_instance == "osd.op_w_latency"`).
		AggregateWindow(`every: 60s, fn: sum, createEmpty: false`).
		Different().
		Derivative(`unit: 1s, nonNegative: true`).
		Group(`columns: ["_time"]`).
		Max(`column: "_value"`).
		Group("").
		String()
}

func (h *helper) genHostStorageUsageRankStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(`fn: (r) => r._measurement == "disk" and r._field == "used_percent" and r.role == "cube"`).
		Group(`columns: ["host"]`).
		Last().
		Keep(`columns: ["host", "_value"]`).
		Top(fmt.Sprintf(`n: %d, columns: ["_value"]`, h.rank.head)).
		String()
}

// note:
// will be supported in the M2
func (h *helper) genHostNetworkIngressRankStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -5m").
		Filter(`fn: (r) => r._measurement == "net" and r.interface =~ /^eth[0-9]+$/ and r.role == "cube" and r._field == "bytes_recv"`).
		AggregateWindow(`every: 1m, fn: sum, createEmpty: false`).
		Derivative(`unit: 1s, nonNegative: true`).
		Map(`fn: (r) => ({ r with used: r._value * 8.0 })`).
		Group(`columns: ["host"]`).
		Max(`column: "used"`).
		Top(fmt.Sprintf(`n: %d, columns: ["used"]`, h.rank.head)).
		String()
}

// note:
// will be supported in the M2
func (h *helper) genHostNetworkEgressRankStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -5m").
		Filter(`fn: (r) => r._measurement == "net" and r.interface =~ /^eth[0-9]+$/ and r.role == "cube" and r._field == "bytes_sent"`).
		AggregateWindow(`every: 1m, fn: sum, createEmpty: false`).
		Derivative(`unit: 1s, nonNegative: true`).
		Map(`fn: (r) => ({ r with used: r._value * 8.0 })`).
		Group(`columns: ["host"]`).
		Max(`column: "used"`).
		Top(fmt.Sprintf(`n: %d, columns: ["used"]`, h.rank.head)).
		String()
}

func (h *helper) genHostCpuUsageHistoryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range(fmt.Sprintf("start: -%s", h.past)).
		Filter(fmt.Sprintf(`fn: (r) => r._measurement == "cpu" and r.host == "%s" and r._field == "usage_idle"`, h.entityId)).
		AggregateWindow(fmt.Sprintf(`every: %s, fn: mean, createEmpty: false`, h.aggregateWindow)).
		Map(`fn: (r) => ({ r with _value: 100.0 - r._value })`).
		Rename(`columns: {_value: "used"}`).
		String()
}

func (h *helper) genHostMemorySizeHistoryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range(fmt.Sprintf("start: -%s", h.past)).
		Filter(fmt.Sprintf(`fn: (r) => r._measurement == "mem" and r.host == "%s" and r._field == "used"`, h.entityId)).
		AggregateWindow(fmt.Sprintf(`every: %s, fn: mean, createEmpty: false`, h.aggregateWindow)).
		Rename(`columns: {_value: "used"}`).
		Keep(`columns: ["_time", "used", "host"]`).
		String()
}

// note:
// will be supported in the M2
func (h *helper) genHostDiskUsageHistoryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(fmt.Sprintf(`fn: (r) => r._measurement == "disk" and r.host == "%s" and r._field == "used_percent"`, h.entityId)).
		Rename(`columns: {_value: "used"}`).
		Keep(`columns: ["_time", "used"]`).
		String()
}

// note:
// will be supported in the M2
func (h *helper) genHostNetworkIngressHistoryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(fmt.Sprintf(`fn: (r) => r._measurement == "net" and r.host == "%s" and r._field == "bytes_recv" and r.interface =~ /^eth[0-9]+$/`, h.entityId)).
		AggregateWindow(`every: 1m, fn: sum`).
		Derivative(`unit: 1s, nonNegative: true`).
		Map(`fn: (r) => ({ r with _value: r._value * 8.0 })`).
		Filter(`fn: (r) => r._value != 0`).
		String()
}

// note:
// will be supported in the M2
func (h *helper) genHostNetworkEgressHistoryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(fmt.Sprintf(`fn: (r) => r._measurement == "net" and r.host == "%s" and r._field == "bytes_sent" and r.interface =~ /^eth[0-9]+$/`, h.entityId)).
		AggregateWindow(`every: 1m, fn: sum`).
		Derivative(`unit: 1s, nonNegative: true`).
		Map(`fn: (r) => ({ r with _value: r._value * 8.0 })`).
		Filter(`fn: (r) => r._value != 0`).
		String()
}

func (h *helper) genVmCpuUsageRankStmt() string {
	query := influx.Query{}
	return query.Bucket("monasca").
		Range("start: -5m").
		Filter(`fn: (r) => r._measurement == "vm.cpu.utilization_norm_perc" and r._field == "value"`).
		Group(`columns: ["resource_id", "vm_name"]`).
		Last().
		Map(`fn: (r) => ({ r with _value: float(v: r._value) })`).
		Group("").
		Sort(`columns: ["_value"], desc: true`).
		Limit(fmt.Sprintf(`n: %d`, h.rank.head)).
		String()
}

func (h *helper) genVmMemoryRankStmt() string {
	query := influx.Query{}
	return query.Bucket("monasca").
		Range("start: -5m").
		Measurement("vm.mem.free_perc").
		Filter(`fn: (r) => r._field == "value"`).
		Group(`columns: ["resource_id", "vm_name"]`).
		Last().
		Map(`fn: (r) => ({ r with used: 100.0 - r._value })`).
		Group(`columns: []`).
		Top(fmt.Sprintf(`n: %d, columns: ["used"]`, h.rank.head)).
		Keep(`columns: ["resource_id", "vm_name", "used", "_time"]`).
		String()
}

func (h *helper) genVmStorageIopsReadRankStmt() string {
	query := influx.Query{}
	return query.Bucket("monasca").
		Range("start: -5m").
		Measurement("vm.io.read_bytes_sec").
		Filter(`fn: (r) => r._field == "value"`).
		Group(`columns: ["resource_id", "vm_name", "device"]`).
		Last().
		Group(`columns: []`).
		Top(fmt.Sprintf(`n: %d, columns: ["_value"]`, h.rank.head)).
		Rename(`columns: {_value: "used"}`).
		Keep(`columns: ["resource_id", "vm_name", "device", "used"]`).
		String()
}

func (h *helper) genVmStorageIopsWriteRankStmt() string {
	query := influx.Query{}
	return query.Bucket("monasca").
		Range("start: -5m").
		Measurement("vm.io.write_bytes_sec").
		Filter(`fn: (r) => r._field == "value"`).
		Group(`columns: ["resource_id", "vm_name", "device"]`).
		Last().
		Group(`columns: []`).
		Top(fmt.Sprintf(`n: %d, columns: ["_value"]`, h.rank.head)).
		Rename(`columns: {_value: "used"}`).
		Keep(`columns: ["resource_id", "vm_name", "device", "used"]`).
		String()
}

func (h *helper) genVmNetworkIngressRankStmt() string {
	query := influx.Query{}
	return query.Bucket("monasca").
		Range("start: -5m").
		Filter(`fn: (r) => r._measurement == "vm.net.in_bytes_sec" and r._field == "value"`).
		Group(`columns: ["resource_id", "vm_name", "device"]`).
		Last().
		Map(`fn: (r) => ({ r with used: r._value * 8.0 })`).
		Group(`columns: []`).
		Top(fmt.Sprintf(`n: %d, columns: ["used"]`, h.rank.head)).
		String()
}

func (h *helper) genVmNetworkEgressRankStmt() string {
	query := influx.Query{}
	return query.Bucket("monasca").
		Range("start: -5m").
		Filter(`fn: (r) => r._measurement == "vm.net.out_bytes_sec" and r._field == "value"`).
		Group(`columns: ["resource_id", "vm_name", "device"]`).
		Last().
		Map(`fn: (r) => ({ r with used: r._value * 8.0 })`).
		Group(`columns: []`).
		Top(fmt.Sprintf(`n: %d, columns: ["used"]`, h.rank.head)).
		String()
}
