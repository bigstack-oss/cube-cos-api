package cubecos

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	log "go-micro.dev/v5/logger"
)

var (
	isMetricTypeValid = map[string]bool{
		"cpuUsage":          true,
		"memoryUsage":       true,
		"diskUsage":         true,
		"diskBandwidth":     true,
		"diskIops":          true,
		"diskReadIops":      true,
		"diskWriteIops":     true,
		"diskLatency":       true,
		"networkTrafficIn":  true,
		"networkTrafficOut": true,
	}

	isResourceTypeValid = map[string]bool{
		"vms":   true,
		"hosts": true,
	}

	isMetricReportTypeValid = map[string]bool{
		"summary": true,
		"history": true,
		"rank":    true,
	}

	hostCpuUsageStmt = `
		from(bucket: "telegraf")
			|> range(start: -2m)
			|> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
			|> aggregateWindow(every: 60s, fn: mean, createEmpty: false)
			|> map(fn: (r) => ({ r with _value: 100.0 - r._value }))
			|> last()
	`

	hostMemoryUsageStmt = `
	    from(bucket: "telegraf")
			|> range(start: -2m)
			|> filter(fn: (r) => r._measurement == "mem" and (r._field == "used" or r._field == "total"))
			|> aggregateWindow(every: 60s, fn: mean, createEmpty: false)
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> map(fn: (r) => ({ r with _value: (r.used * 100.0) / r.total }))
			|> last()
	`

	hostMemoryUsageRankStmt = `
		from(bucket: "telegraf")
			|> range(start: -2m)
			|> filter(fn: (r) => 
				r._measurement == "mem" and
				r._field == "used_percent" and
				r.role == "cube"
			)
			|> last()
			|> group(columns: ["host"])
			|> top(n: 10, columns: ["_value"])
	`

	hostCpuUsageRankStmt = `
		from(bucket: "telegraf")
			|> range(start: -2m)
			|> filter(fn: (r) => 
				r._measurement == "cpu" and
				r._field == "usage_idle" and
				r.role == "cube"
			)
			|> group(columns: ["host"])
			|> last()
			|> map(fn: (r) => ({ r with used: 100.0 - r._value }))
			|> group()
			|> top(n: %d, columns: ["used"])
			|> keep(columns: ["host", "used"])
	`

	hostStorageReadBandwidthStmt = `
		from(bucket: "ceph")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => 
				r._measurement == "ceph_daemon_stats" and
				r.ceph_daemon =~ /^osd\.[0-9]+$/ and
				r.type_instance == "osd.op_r_out_bytes"
			)
			|> aggregateWindow(every: 60s, fn: sum, createEmpty: false)
			|> derivative(unit: 1s, nonNegative: true)
			|> group(columns: ["_time"])
			|> max(column: "_value")
			|> group()
	`

	hostStorageWriteBandwidthStmt = `
		from(bucket: "ceph")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => 
				r._measurement == "ceph_daemon_stats" and
				r.ceph_daemon =~ /^osd\.[0-9]+$/ and
				r.type_instance == "osd.op_w_in_bytes"
			)
			|> aggregateWindow(every: 60s, fn: sum, createEmpty: false)
			|> derivative(unit: 1s, nonNegative: true)
			|> group(columns: ["_time"])
			|> max(column: "_value")
			|> group()
	`

	hostStorageReadIopsStmt = `
		from(bucket: "ceph")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => 
				r._measurement == "ceph_daemon_stats" and
				r.ceph_daemon =~ /^osd\.[0-9]+$/ and
				r.type_instance == "osd.op_r"
			)
			|> aggregateWindow(every: 60s, fn: sum, createEmpty: false)
			|> derivative(unit: 1s, nonNegative: true)
			|> group(columns: ["_time"])
			|> max(column: "_value")
			|> group()
	`

	hostStorageWriteIopsStmt = `
		from(bucket: "ceph")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => 
				r._measurement == "ceph_daemon_stats" and
				r.ceph_daemon =~ /^osd\.[0-9]+$/ and
				r.type_instance == "osd.op_w"
			)
			|> aggregateWindow(every: 60s, fn: sum, createEmpty: false)
			|> derivative(unit: 1s, nonNegative: true)
			|> group(columns: ["_time"])
			|> max(column: "_value")
			|> group()
	`

	hostStorageReadLatencyStmt = `
		from(bucket: "ceph")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => 
				r._measurement == "ceph_daemon_stats" and
				r.ceph_daemon =~ /^osd\.[0-9]+$/ and
				r.type_instance == "osd.op_r_latency"
			)
			|> aggregateWindow(every: 60s, fn: sum, createEmpty: false)
			|> difference()
			|> derivative(unit: 1s, nonNegative: true)
			|> group(columns: ["_time"])
			|> max(column: "_value")
			|> group()
	`

	hostStorageWriteLatencyStmt = `
		from(bucket: "ceph")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => 
				r._measurement == "ceph_daemon_stats" and
				r.ceph_daemon =~ /^osd\.[0-9]+$/ and
				r.type_instance == "osd.op_w_latency"
			)
			|> aggregateWindow(every: 60s, fn: sum, createEmpty: false)
			|> difference()
			|> derivative(unit: 1s, nonNegative: true)
			|> group(columns: ["_time"])
			|> max(column: "_value")
			|> group()
	`

	hostStorageUsageRankStmt = `
		from(bucket: "telegraf")
			|> range(start: -2m)
			|> filter(fn: (r) => 
				r._measurement == "disk" and 
				r._field == "used_percent" and 
				r.role == "cube"
			)
			|> group(columns: ["host"])
			|> last()
			|> keep(columns: ["host", "_value"])
			|> top(n: %d, columns: ["_value"])
	`

	hostNetworkIngressRankStmt = `
		from(bucket: "telegraf")
			|> range(start: -5m)
			|> filter(fn: (r) => 
				r._measurement == "net" and
				r.interface =~ /^eth[0-9]+$/ and
				r.role == "cube" and
				r._field == "bytes_recv"
			)
			|> aggregateWindow(every: 1m, fn: sum, createEmpty: false)
			|> derivative(unit: 1s, nonNegative: true)
			|> map(fn: (r) => ({ r with used: r._value * 8.0 }))
			|> group(columns: ["host"])
			|> max(column: "used")
			|> top(n: 10, columns: ["used"])
	`

	hostNetworkEgressRankStmt = `
		from(bucket: "telegraf")
			|> range(start: -5m)
			|> filter(fn: (r) => 
				r._measurement == "net" and
				r.interface =~ /^eth[0-9]+$/ and
				r.role == "cube" and
				r._field == "bytes_sent"
			)
			|> aggregateWindow(every: 1m, fn: sum, createEmpty: false)
			|> derivative(unit: 1s, nonNegative: true)
			|> map(fn: (r) => ({ r with used: r._value * 8.0 }))
			|> group(columns: ["host"])
			|> max(column: "used")
			|> top(n: 10, columns: ["used"])
	`

	vmCpuUsageRankStmt = `
		from(bucket: "monasca")
			|> range(start: -5m)
			|> filter(fn: (r) => 
				r._measurement == "vm.cpu.utilization_norm_perc" and 
				r._field == "value"
			)
			|> group(columns: ["resource_id", "vm_name"])
			|> last()
			|> map(fn: (r) => ({ r with _value: float(v: r._value) }))
			|> group()
			|> sort(columns: ["_value"], desc: true)
			|> limit(n: %d)
	`

	vmMemoryRankStmt = `
		from(bucket: "monasca")
			|> range(start: -5m)
			|> filter(fn: (r) => r._measurement == "vm.mem.free_perc")
			|> filter(fn: (r) => r._field == "value")
			|> group(columns: ["resource_id", "vm_name"])
			|> last()
			|> map(fn: (r) => ({ r with used: 100.0 - r._value }))
			|> group(columns: [])
			|> top(n: %d, columns: ["used"])
			|> keep(columns: ["resource_id", "vm_name", "used", "_time"])
	`

	vmStorageIopsReadRankStmt = `
		from(bucket: "monasca")
			|> range(start: -5m)
			|> filter(fn: (r) => r._measurement == "vm.io.read_bytes_sec")
			|> filter(fn: (r) => r._field == "value")
			|> group(columns: ["resource_id", "vm_name", "device"])
			|> last()
			|> group(columns: [])
			|> top(n: %d, columns: ["_value"])
			|> rename(columns: {_value: "used"})
			|> keep(columns: ["resource_id", "vm_name", "device", "used"])
	`

	vmStorageIopsWriteRankStmt = `
		from(bucket: "monasca")
			|> range(start: -5m)
			|> filter(fn: (r) => r._measurement == "vm.io.write_bytes_sec")
			|> filter(fn: (r) => r._field == "value")
			|> group(columns: ["resource_id", "vm_name", "device"])
			|> last()
			|> group(columns: [])
			|> top(n: %d, columns: ["_value"])
			|> rename(columns: {_value: "used"})
			|> keep(columns: ["resource_id", "vm_name", "device", "used"])
`

	vmNetworkIngressRankStmt = `
		from(bucket: "monasca")
			|> range(start: -5m)
			|> filter(fn: (r) => 
				r._measurement == "vm.net.in_bytes_sec" and
				r._field == "value"
			)
			|> group(columns: ["resource_id", "vm_name", "device"])
			|> last()
			|> map(fn: (r) => ({ r with used: r._value * 8.0 }))
			|> group(columns: [])
			|> top(n: %d, columns: ["used"])
	`

	vmNetworkEgressRankStmt = `
		from(bucket: "monasca")
			|> range(start: -5m)
			|> filter(fn: (r) => 
				r._measurement == "vm.net.out_bytes_sec" and
				r._field == "value"
			)
			|> group(columns: ["resource_id", "vm_name", "device"])
			|> last()
			|> map(fn: (r) => ({ r with used: r._value * 8.0 }))
			|> group(columns: [])
			|> top(n: %d, columns: ["used"])
	`
)

func IsMetricTypeValid(t string) bool {
	return isMetricTypeValid[t]
}

func IsEntityTypeValid(t string) bool {
	return isResourceTypeValid[t]
}

func IsMetricReportTypeValid(t string) bool {
	return isMetricReportTypeValid[t]
}

func GetDataCenterSummary() (*Summary, error) {
	host, err := GetHostSummary()
	if err != nil {
		log.Errorf("failed to get host summary: %v", err)
		return nil, err
	}

	vm, err := GetVmSummary()
	if err != nil {
		log.Errorf("failed to get vm summary: %v", err)
		return nil, err
	}

	return &Summary{Host: *host, Vm: *vm}, nil
}

func GetHostSummary() (*HostSummary, error) {
	role, err := GetRoleStatus()
	if err != nil {
		log.Errorf("failed to get role status: %v", err)
		return nil, err
	}

	cpu, err := GetCpuSummaryOfHosts()
	if err != nil {
		log.Errorf("failed to get host cpu summary: %v", err)
		return nil, err
	}

	memory, err := GetMemoryUsageSummaryOfHosts()
	if err != nil {
		log.Errorf("failed to get host memory summary: %v", err)
		return nil, err
	}

	return &HostSummary{
		Role: *role,
		Usage: definition.Usage{
			Vcpu:   *cpu,
			Memory: *memory,
		},
	}, nil
}

func GetVmSummary() (*VmSummary, error) {
	status, err := GetVmStatus()
	if err != nil {
		return nil, err
	}

	usage, err := GetVmUsage()
	if err != nil {
		return nil, err
	}

	return &VmSummary{
		Status: *status,
		Usage:  *usage,
	}, nil
}

func GetCpuSummaryOfHosts() (*definition.ComputeStatistic, error) {
	c, cancel, err := influx.GetQueryCursor(hostCpuUsageStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageOfHost(c)
}

func GetCpuUsageRankOfHosts(top int) ([]definition.HostPercentageUsage, error) {
	stmt := fmt.Sprintf(hostCpuUsageRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageRankOfHost(c)
}

func GetMemoryUsageSummaryOfHosts() (*definition.SpaceStatistic, error) {
	c, cancel, err := influx.GetQueryCursor(hostMemoryUsageStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageSummaryOfHost(c)
}

func GetMemoryUsageRankOfHosts() ([]definition.HostPercentageUsage, error) {
	c, cancel, err := influx.GetQueryCursor(hostMemoryUsageRankStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageRankOfHost(c)
}

func GetDiskStorageBandwidthHistory(period definition.Period) (*definition.StorageBandwidthSeries, error) {
	stmt := fmt.Sprintf(hostStorageReadBandwidthStmt, period.Start, period.Stop)
	read, err := getDiskBandwidthHistory(stmt)
	if err != nil {
		log.Errorf("failed to get host storage read bandwidth series: %v", err)
		return nil, err
	}

	stmt = fmt.Sprintf(hostStorageWriteBandwidthStmt, period.Start, period.Stop)
	write, err := getDiskBandwidthHistory(stmt)
	if err != nil {
		log.Errorf("failed to get host storage write bandwidth series: %v", err)
		return nil, err
	}

	return &definition.StorageBandwidthSeries{
		Read:  read,
		Write: write,
	}, nil
}

func GetDiskIopsHistoryOfHosts(period definition.Period) (*definition.StorageIopsSeries, error) {
	readStmt := fmt.Sprintf(hostStorageReadIopsStmt, period.Start, period.Stop)
	readSeries, err := getDiskIopsHistoryOfHosts(readStmt)
	if err != nil {
		return nil, err
	}

	writeStmt := fmt.Sprintf(hostStorageWriteIopsStmt, period.Start, period.Stop)
	writeSeries, err := getDiskIopsHistoryOfHosts(writeStmt)
	if err != nil {
		return nil, err
	}

	return &definition.StorageIopsSeries{
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func GeDiskLatencyHistoryOfHosts(period definition.Period) (*definition.StorageLatencySeries, error) {
	readStmt := fmt.Sprintf(hostStorageReadLatencyStmt, period.Start, period.Stop)
	readSeries, err := geDiskLatencyHistoryOfHosts(readStmt)
	if err != nil {
		return nil, err
	}

	writeStmt := fmt.Sprintf(hostStorageWriteLatencyStmt, period.Start, period.Stop)
	writeSeries, err := geDiskLatencyHistoryOfHosts(writeStmt)
	if err != nil {
		return nil, err
	}

	return &definition.StorageLatencySeries{
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func getDiskIopsHistoryOfHosts(stmt string) ([]definition.TimeOpsPoint, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func geDiskLatencyHistoryOfHosts(stmt string) ([]definition.TimeLatencyPoint, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskLatencyHistory(c)
}

func GetDiskUsageRankOfHosts(top int) ([]definition.HostPercentageUsage, error) {
	stmt := fmt.Sprintf(hostStorageUsageRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseHostStorageUsageRank(c)
}

func GetNetworkTrafficInRankOfHosts() ([]definition.HostNetworkPacket, error) {
	c, cancel, err := influx.GetQueryCursor(hostNetworkIngressRankStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return hostNetworkIngressRank(c)
}

func GetNetworkTrafficOutRankOfHosts() ([]definition.HostNetworkPacket, error) {
	c, cancel, err := influx.GetQueryCursor(hostNetworkEgressRankStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return hostNetworkIngressRank(c)
}

func GetCpuUsageRankOfVms(top int) ([]definition.VmPercentageUsage, error) {
	stmt := fmt.Sprintf(vmCpuUsageRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseVmCpuUsageRank(c)
}

func GetMemoryUsageRankOfVms(top int) ([]definition.VmMetricsUsage, error) {
	stmt := fmt.Sprintf(vmMemoryRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseVmMemoryRank(c)
}

func GetDiskReadIopsRankOfVms(top int) ([]definition.VmMetricsUsage, error) {
	stmt := fmt.Sprintf(vmStorageIopsReadRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parsDiskIopsRankOfVm(c)
}

func GetDiskWriteIopsRankOfVms(top int) ([]definition.VmMetricsUsage, error) {
	stmt := fmt.Sprintf(vmStorageIopsWriteRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parsDiskIopsRankOfVm(c)
}

func GetNetworkTrafficInRankOfVms(top int) ([]definition.VmMetricsUsage, error) {
	stmt := fmt.Sprintf(vmNetworkIngressRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseVmNetworkPacketRank(c)
}

func GetNetworkTrafficOutRankOfVms(top int) ([]definition.VmMetricsUsage, error) {
	stmt := fmt.Sprintf(vmNetworkEgressRankStmt, top)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseVmNetworkPacketRank(c)
}

func parseCpuUsageRankOfHost(c *api.QueryTableResult) ([]definition.HostPercentageUsage, error) {
	rank := []definition.HostPercentageUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.HostPercentageUsage{
				Id:          parseHost(c.Record()),
				Name:        parseHost(c.Record()),
				UsedPercent: parseCpuUsedOfHost(c.Record()),
				FreePercent: 100 - parseCpuUsedOfHost(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseVmCpuUsageRank(c *api.QueryTableResult) ([]definition.VmPercentageUsage, error) {
	rank := []definition.VmPercentageUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.VmPercentageUsage{
				Id:          parseResourceId(c.Record()),
				Name:        parseVmName(c.Record()),
				UsedPercent: math.RoundDown(c.Record().Value().(float64), 4),
				FreePercent: 100 - math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseCpuUsageOfHost(c *api.QueryTableResult) (*definition.ComputeStatistic, error) {
	usage := definition.ComputeStatistic{}
	for c.Next() {
		record := c.Record()
		usage = parseCpuUsage(record)
		break
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return &usage, nil
}

func parseCpuUsage(record *query.FluxRecord) definition.ComputeStatistic {
	usedPercent := record.Value().(float64)
	return definition.ComputeStatistic{
		UsedPercent: math.RoundDown(usedPercent, 4),
		FreePercent: math.RoundDown(100-usedPercent, 4),
	}
}

func parseMemoryUsageRankOfHost(c *api.QueryTableResult) ([]definition.HostPercentageUsage, error) {
	rank := []definition.HostPercentageUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.HostPercentageUsage{
				Id:          parseHost(c.Record()),
				Name:        parseHost(c.Record()),
				UsedPercent: parseUsed(c.Record()),
				FreePercent: 100 - parseUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseVmMemoryRank(c *api.QueryTableResult) ([]definition.VmMetricsUsage, error) {
	rank := []definition.VmMetricsUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.VmMetricsUsage{
				Id:          parseResourceId(c.Record()),
				Name:        parseVmName(c.Record()),
				UsedPercent: parseCpuUsedOfVm(c.Record()),
				FreePercent: math.RoundDown(100-parseCpuUsedOfVm(c.Record()), 4),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseMemoryUsageSummaryOfHost(c *api.QueryTableResult) (*definition.SpaceStatistic, error) {
	memoryUsage := definition.SpaceStatistic{}
	for c.Next() {
		record := c.Record()
		memoryUsage = parseMemoryUsage(record)
		break
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return &memoryUsage, nil
}

func parseMemoryUsage(record *query.FluxRecord) definition.SpaceStatistic {
	usedPercent := record.Value().(float64)
	return definition.SpaceStatistic{
		UsedPercent: math.RoundDown(usedPercent, 4),
		FreePercent: math.RoundDown(100-usedPercent, 4),
	}
}

func parseHostStorageUsageRank(c *api.QueryTableResult) ([]definition.HostPercentageUsage, error) {
	rank := []definition.HostPercentageUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.HostPercentageUsage{
				Id:          parseHost(c.Record()),
				Name:        parseHost(c.Record()),
				UsedPercent: parseUsed(c.Record()),
				FreePercent: 100 - parseUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseDiskOpsHistory(c *api.QueryTableResult) ([]definition.TimeOpsPoint, error) {
	points := []definition.TimeOpsPoint{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeOpsPoint{
				Time: date.Format(time.RFC3339),
				Ops:  math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parseDiskLatencyHistory(c *api.QueryTableResult) ([]definition.TimeLatencyPoint, error) {
	points := []definition.TimeLatencyPoint{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeLatencyPoint{
				Time: definition.TimeLocalISO8601(date),
				Ms:   math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parsDiskIopsRankOfVm(c *api.QueryTableResult) ([]definition.VmMetricsUsage, error) {
	rank := []definition.VmMetricsUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.VmMetricsUsage{
				Id:     parseResourceId(c.Record()),
				Name:   parseVmName(c.Record()),
				Device: parseDevice(c.Record()),
				Usage:  parseStorageUsedOfVm(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func getDiskBandwidthHistory(stmt string) ([]definition.TimeBytesPoint, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskBandwidthHistory(c)
}

func parseDiskBandwidthHistory(c *api.QueryTableResult) ([]definition.TimeBytesPoint, error) {
	points := []definition.TimeBytesPoint{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeBytesPoint{
				Time:  definition.TimeLocalISO8601(date),
				Bytes: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func hostNetworkIngressRank(c *api.QueryTableResult) ([]definition.HostNetworkPacket, error) {
	rank := []definition.HostNetworkPacket{}
	for c.Next() {
		rank = append(
			rank,
			definition.HostNetworkPacket{
				Id:      parseHost(c.Record()),
				Name:    parseHost(c.Record()),
				Packets: parseUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseVmNetworkPacketRank(c *api.QueryTableResult) ([]definition.VmMetricsUsage, error) {
	rank := []definition.VmMetricsUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.VmMetricsUsage{
				Id:    parseResourceId(c.Record()),
				Name:  parseVmName(c.Record()),
				Usage: parseUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseHost(record *query.FluxRecord) string {
	host, ok := record.ValueByKey("host").(string)
	if !ok {
		return "unknown host"
	}

	return host
}

func parseCpuUsedOfHost(record *query.FluxRecord) float64 {
	used, ok := record.ValueByKey("used").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseCpuUsedOfVm(record *query.FluxRecord) float64 {
	used, ok := record.ValueByKey("used").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseStorageUsedOfVm(record *query.FluxRecord) float64 {
	used, ok := record.ValueByKey("used").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseUsed(record *query.FluxRecord) float64 {
	used, ok := record.Value().(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseResourceId(record *query.FluxRecord) string {
	id, ok := record.ValueByKey("resource_id").(string)
	if !ok {
		return "unknown id"
	}

	return id
}

func parseVmName(record *query.FluxRecord) string {
	name, ok := record.ValueByKey("vm_name").(string)
	if !ok {
		return "unknown name"
	}

	return name
}

func parseDevice(record *query.FluxRecord) string {
	device, ok := record.ValueByKey("device").(string)
	if !ok {
		return "unknown device"
	}

	return device
}
