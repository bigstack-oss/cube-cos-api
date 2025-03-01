package cubecos

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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

	isEntityTypeValid = map[string]bool{
		"vms":   true,
		"vm":    true,
		"hosts": true,
		"host":  true,
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

	hostCpuUsageHistoryStmt = `
		from(bucket: "telegraf")
			|> range(start: -1h)
			|> filter(fn: (r) => r._measurement == "cpu" and r.host == "%s" and r._field == "usage_idle")
			|> map(fn: (r) => ({ r with _value: 100.0 - r._value }))
			|> rename(columns: {_value: "used"})
	`

	hostMemoryUsageHistoryStmt = `
		from(bucket: "telegraf")
			|> range(start: -1h)
			|> filter(fn: (r) => r._measurement == "mem" and r.host == "%s" and r._field == "used_percent")
			|> rename(columns: {_value: "used"})
			|> keep(columns: ["_time", "used", "host"])
	`

	hostDiskUsageHistoryStmt = `
		from(bucket: "telegraf")
			|> range(start: -1h)
			|> filter(fn: (r) => r._measurement == "disk" and r.host == "%s" and r._field == "used_percent")
			|> rename(columns: {_value: "used"})
			|> keep(columns: ["_time", "used"])
	`

	hostNetworkIngressHistoryStmt = `
		from(bucket: "telegraf")
			|> range(start: -1h)
			|> filter(fn: (r) => 
				r._measurement == "net" and
				r.host == "%s" and
				r._field == "bytes_recv" and
				r.interface =~ /^eth[0-9]+$/
			)
			|> aggregateWindow(every: 1m, fn: sum)
			|> derivative(unit: 1s, nonNegative: true)
			|> map(fn: (r) => ({ r with _value: r._value * 8.0 }))
			|> filter(fn: (r) => r._value != 0)
	`

	hostNetworkEgressHistoryStmt = `
		from(bucket: "telegraf")
			|> range(start: -1h)
			|> filter(fn: (r) => 
				r._measurement == "net" and
				r.host == "%s" and
				r._field == "bytes_sent" and
				r.interface =~ /^eth[0-9]+$/
			)
			|> aggregateWindow(every: 1m, fn: sum)
			|> derivative(unit: 1s, nonNegative: true)
			|> map(fn: (r) => ({ r with _value: r._value * 8.0 }))
			|> filter(fn: (r) => r._value != 0)
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

	vmCpuUsageHistoryStmt = `
		from(bucket: "monasca")
			|> range(start: -1h)
			|> filter(fn: (r) =>
				r._measurement == "vm.cpu.utilization_norm_perc" and
				r.resource_id == "%s" and
				r._field == "value"
			)
	`

	vmMemoryUsageHistoryStmt = `
		from(bucket: "monasca")
            |> range(start: -1h)
            |> filter(fn: (r) => 
                r._measurement == "vm.mem.free_perc" and
                r.resource_id == "%s" and
                r._field == "value"
            )
            |> map(fn: (r) => ({ 
                r with _value: 100.0 - r._value 
            }))
	`

	vmStorageIopsReadHistoryStmt = `
		from(bucket: "monasca")
			|> range(start: -1h)
			|> filter(fn: (r) =>
				r._measurement == "vm.io.read_bytes_sec" and
				r.resource_id == "%s" and
				r.device == "%s" and
				r._field == "value"
			)
	`

	vmStorageIopsWriteHistoryStmt = `
		from(bucket: "monasca")
			|> range(start: -1h)
			|> filter(fn: (r) =>
				r._measurement == "vm.io.write_bytes_sec" and
				r.resource_id == "%s" and
				r.device == "%s" and
				r._field == "value"
			)
	`

	vmNetworkIngressHistoryStmt = `
		from(bucket: "monasca")
			|> range(start: -1h)
			|> filter(fn: (r) =>
				r._measurement == "vm.net.in_bytes_sec" and
				r.resource_id == "%s" and
				r.device == "%s" and
				r._field == "value"
			)
			|> map(fn: (r) => ({ r with _value: r._value * 8.0 }))
	`

	vmNetworkEgressHistoryStmt = `
		from(bucket: "monasca")
			|> range(start: -1h)
			|> filter(fn: (r) => 
				r._measurement == "vm.net.out_bytes_sec" and
				r.resource_id == "%s" and
				r.device == "%s" and
				r._field == "value"
			)
			|> map(fn: (r) => ({ r with _value: r._value * 8.0 }))
	`
)

func IsMetricTypeValid(t string) bool {
	return isMetricTypeValid[t]
}

func IsEntityTypeValid(t string) bool {
	return isEntityTypeValid[t]
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

	dataCenter, err := GetDataCenterUsage(host)
	if err != nil {
		log.Errorf("failed to get data center usage: %v", err)
		return nil, err
	}

	vm, err := GetVmSummary()
	if err != nil {
		log.Errorf("failed to get vm summary: %v", err)
		return nil, err
	}

	return &Summary{
		DataCenter: *dataCenter,
		Host:       *host,
		Vm:         *vm,
	}, nil
}

func GetDataCenterUsage(hostSummary *HostSummary) (*DataCenterSummary, error) {
	return &DataCenterSummary{
		Usage: definition.DataCenterUsage{
			Cpu:    GetCpuAverageOfHosts(hostSummary.ListCpuUsages()),
			Memory: GetMemoryAverageOfHosts(hostSummary.ListMemoryUsages()),
		},
	}, nil
}

func GetCpuAverageOfHosts(cpuStats []definition.ComputeStatistic) definition.ComputeStatistic {
	totalCores := float64(0)
	usedCores := float64(0)
	usedPercent := float64(0)
	freeCores := float64(0)
	freePercent := float64(0)

	for _, stat := range cpuStats {
		totalCores += stat.TotalCores
		usedCores += stat.UsedCores
		usedPercent += stat.UsedPercent
		freeCores += stat.FreeCores
		freePercent += stat.FreePercent
	}

	return definition.ComputeStatistic{
		TotalCores:  totalCores,
		UsedCores:   math.RoundDown(usedCores/float64(len(cpuStats)), 4),
		UsedPercent: math.RoundDown(usedPercent/float64(len(cpuStats)), 4),
		FreeCores:   math.RoundDown(freeCores/float64(len(cpuStats)), 4),
		FreePercent: math.RoundDown(freePercent/float64(len(cpuStats)), 4),
	}
}

func GetMemoryAverageOfHosts(spaceStats []definition.SpaceStatistic) definition.SpaceStatistic {
	totalMiB := float64(0)
	usedMiB := float64(0)
	usedPercent := float64(0)
	freeMiB := float64(0)
	freePercent := float64(0)

	for _, stat := range spaceStats {
		totalMiB += stat.TotalMiB
		usedMiB += stat.UsedMiB
		usedPercent += stat.UsedPercent
		freeMiB += stat.FreeMiB
		freePercent += stat.FreePercent
	}

	return definition.SpaceStatistic{
		TotalMiB:    totalMiB / float64(len(spaceStats)),
		UsedMiB:     usedMiB / float64(len(spaceStats)),
		UsedPercent: usedPercent / float64(len(spaceStats)),
		FreeMiB:     freeMiB / float64(len(spaceStats)),
		FreePercent: freePercent / float64(len(spaceStats)),
	}
}

func GetHostSummary() (*HostSummary, error) {
	roleStatus, err := GetRoleStatus()
	if err != nil {
		log.Errorf("failed to get role status: %v", err)
		return nil, err
	}

	nodes, err := definition.ListNodes()
	if err != nil {
		log.Errorf("failed to list nodes: %v", err)
		return nil, err
	}

	host := &HostSummary{Role: *roleStatus}
	for _, node := range nodes {
		usage, err := GetHostUsage(node)
		if err != nil {
			continue
		}

		host.Usages = append(
			host.Usages,
			HostUsage{
				Role:      node.Role,
				Name:      node.Hostname,
				Address:   node.Address,
				HostUsage: *usage,
			},
		)
	}

	return host, nil
}

func GetHostUsage(node *definition.Node) (*definition.HostUsage, error) {
	cpuStat, err := GetCpuSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("failed to get cpu summary of host %s: %v", node.Hostname, err)
		return nil, err
	}

	memoryStat, err := GetMemoryUsageSummaryOfHost(node.Hostname)
	if err != nil {
		log.Errorf("failed to get memory summary of host %s: %v", node.Hostname, err)
		return nil, err
	}

	return &definition.HostUsage{
		Cpu:    *cpuStat,
		Memory: *memoryStat,
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
		Status:  *status,
		VmUsage: *usage,
	}, nil
}

func GetCpuSummaryOfHosts(stmt string) (*definition.ComputeStatistic, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageOfHost(c)
}

func GetCpuSummaryOfHost(hostname string) (*definition.ComputeStatistic, error) {
	if !definition.IsLocalNode(hostname) {
		return askTheHostForCpuSummary(hostname)
	}

	usagePerCore, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	totalCores := float64(len(usagePerCore))
	usedCores := float64(0)
	for _, usage := range usagePerCore {
		usedCores += (usage / 100.0)
	}

	freeCores := totalCores - usedCores
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	return &definition.ComputeStatistic{
		TotalCores:  float64(runtime.NumCPU()),
		UsedCores:   math.RoundDown(usedCores, 4),
		UsedPercent: math.RoundDown(cpuPercents[0], 4),
		FreeCores:   math.RoundDown(freeCores, 4),
		FreePercent: math.RoundDown(100-cpuPercents[0], 4),
	}, nil
}

func GetCpuHistoryOfHost(stmt string) ([]definition.TimeUsedPercent, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageHistory(c)
}

func parseCpuUsageHistory(c *api.QueryTableResult) ([]definition.TimeUsedPercent, error) {
	points := []definition.TimeUsedPercent{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeUsedPercent{
				Time:        definition.TimeLocalISO8601(date),
				UsedPercent: parseUsedOfHost(c.Record()),
			},
		)
	}

	return points, nil
}

func askTheHostForCpuSummary(hostname string) (*definition.ComputeStatistic, error) {
	node, err := definition.GetNodeByHostname(hostname)
	if err != nil {
		return nil, err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&definition.ComputeStatistic{}).
		SetHeader("Authorization", node.GetBearerToken()).
		Get(node.GetMetricUrl("cpuUsage", "summary"))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"failed to get cpu usage of host %s: %d %s",
			hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	return resp.Result().(*definition.ComputeStatistic), nil
}

func GetCpuUsageRankOfHosts(stmt string) ([]definition.HostPercentageUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseCpuUsageRankOfHost(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToCpuUsageRank(rank)
	return rank, nil
}

func appendHistoryToCpuUsageRank(rank []definition.HostPercentageUsage) {
	for i, host := range rank {
		history, err := GetCpuHistoryOfHost(genHostCpuUsageHistoryStmt(host.Id))
		if err != nil {
			log.Errorf("failed to get cpu history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetMemoryUsageSummaryOfHosts() (*definition.SpaceStatistic, error) {
	nodes, err := definition.ListNodes()
	if err != nil {
		log.Errorf("failed to list nodes: %v", err)
		return nil, err
	}

	usages := []definition.SpaceStatistic{}
	for _, node := range nodes {
		usage, err := GetMemoryUsageSummaryOfHost(node.Hostname)
		if err != nil {
			continue
		}

		usages = append(usages, *usage)
	}

	stat := GetMemoryAverageOfHosts(usages)
	return &stat, nil
}

func GetMemoryUsageSummaryOfHost(hostname string) (*definition.SpaceStatistic, error) {
	if !definition.IsLocalNode(hostname) {
		return askTheHostForMemorySummary(hostname)
	}

	stat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	usedPercent := stat.UsedPercent
	freePercent := 100.0 - usedPercent
	return &definition.SpaceStatistic{
		TotalMiB:    math.RoundDown(float64(stat.Total)/1024/1024, 4),
		UsedMiB:     math.RoundDown(float64(stat.Used)/1024/1024, 4),
		FreeMiB:     math.RoundDown(float64(stat.Total-stat.Used)/1024/1024, 4),
		UsedPercent: math.RoundDown(usedPercent, 4),
		FreePercent: math.RoundDown(freePercent, 4),
	}, nil
}

func askTheHostForMemorySummary(hostname string) (*definition.SpaceStatistic, error) {
	node, err := definition.GetNodeByHostname(hostname)
	if err != nil {
		return nil, err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&definition.SpaceStatistic{}).
		SetHeader("Authorization", node.GetBearerToken()).
		Get(node.GetMetricUrl("memoryUsage", "summary"))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"failed to get memory usage of host %s: %d %s",
			hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	return resp.Result().(*definition.SpaceStatistic), nil
}

func GetMemoryUsageRankOfHosts(stmt string) ([]definition.HostPercentageUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseMemoryUsageRankOfHost(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToMemoryUsageRank(rank)
	return rank, nil
}

func appendHistoryToMemoryUsageRank(rank []definition.HostPercentageUsage) {
	for i, host := range rank {
		history, err := GetMemoryHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get memory history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetMemoryHistoryOfHost(entityId string, period definition.Period) ([]definition.TimeUsedPercent, error) {
	stmt := fmt.Sprintf(hostMemoryUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageHistory(c)
}

func parseMemoryUsageHistory(c *api.QueryTableResult) ([]definition.TimeUsedPercent, error) {
	points := []definition.TimeUsedPercent{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeUsedPercent{
				Time:        definition.TimeLocalISO8601(date),
				UsedPercent: parseUsedOfHost(c.Record()),
			},
		)
	}

	return points, nil
}

func GetDiskStorageBandwidthHistory(readStmt, writeStmt string) (*definition.StorageBandwidthSeries, error) {
	read, err := getDiskBandwidthHistory(readStmt)
	if err != nil {
		log.Errorf("failed to get host storage read bandwidth series: %v", err)
		return nil, err
	}

	write, err := getDiskBandwidthHistory(writeStmt)
	if err != nil {
		log.Errorf("failed to get host storage write bandwidth series: %v", err)
		return nil, err
	}

	return &definition.StorageBandwidthSeries{
		Read:  read,
		Write: write,
	}, nil
}

func GetDiskIopsHistoryOfHosts(readStmt, writeStmt string) (*definition.StorageIopsSeries, error) {
	readSeries, err := getDiskIopsHistoryOfHosts(readStmt)
	if err != nil {
		return nil, err
	}

	writeSeries, err := getDiskIopsHistoryOfHosts(writeStmt)
	if err != nil {
		return nil, err
	}

	return &definition.StorageIopsSeries{
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func GeDiskLatencyHistoryOfHosts(readStmt, writeStmt string) (*definition.StorageLatencySeries, error) {
	readSeries, err := geDiskLatencyHistoryOfHosts(readStmt)
	if err != nil {
		return nil, err
	}

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

func geDiskLatencyHistoryOfHosts(stmt string) ([]definition.TimeMillisecondPoint, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskLatencyHistory(c)
}

func GetDiskUsageRankOfHosts(stmt string) ([]definition.HostPercentageUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseHostStorageUsageRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToDiskUsageRank(rank)
	return rank, nil
}

func appendHistoryToDiskUsageRank(rank []definition.HostPercentageUsage) {
	for i, host := range rank {
		history, err := GetDiskUsageHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get disk usage history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskUsageHistoryOfHost(entityId string, period definition.Period) ([]definition.TimeUsedPercent, error) {
	stmt := fmt.Sprintf(hostDiskUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskUsageHistory(c)
}

func parseDiskUsageHistory(c *api.QueryTableResult) ([]definition.TimeUsedPercent, error) {
	points := []definition.TimeUsedPercent{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeUsedPercent{
				Time:        definition.TimeLocalISO8601(date),
				UsedPercent: parseUsedOfHost(c.Record()),
			},
		)
	}

	return points, nil
}

func GetNetworkTrafficInRankOfHosts() ([]definition.HostNetworkPacket, error) {
	c, cancel, err := influx.GetQueryCursor(hostNetworkIngressRankStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := hostNetworkIngressRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToNetworkTrafficInRank(rank)
	return rank, nil
}

func appendHistoryToNetworkTrafficInRank(rank []definition.HostNetworkPacket) {
	for i, host := range rank {
		history, err := GetNetworkTrafficInHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get network traffic in history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficInHistoryOfHost(entityId string, period definition.Period) ([]definition.TimePacketsPoint, error) {
	stmt := fmt.Sprintf(hostNetworkIngressHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func parseNetworkTrafficHistory(c *api.QueryTableResult) ([]definition.TimePacketsPoint, error) {
	points := []definition.TimePacketsPoint{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimePacketsPoint{
				Time:    definition.TimeLocalISO8601(date),
				Packets: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func GetNetworkTrafficOutRankOfHosts() ([]definition.HostNetworkPacket, error) {
	c, cancel, err := influx.GetQueryCursor(hostNetworkEgressRankStmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := hostNetworkIngressRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToNetworkTrafficOutRank(rank)
	return rank, nil
}

func appendHistoryToNetworkTrafficOutRank(rank []definition.HostNetworkPacket) {
	for i, host := range rank {
		history, err := GetNetworkTrafficOutHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get network traffic out history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficOutHistoryOfHost(entityId string, period definition.Period) ([]definition.TimePacketsPoint, error) {
	stmt := fmt.Sprintf(hostNetworkEgressHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func GetCpuUsageRankOfVms(stmt string) ([]definition.VmPercentageUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseVmCpuUsageRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToCpuUsageRankOfVm(rank)
	return rank, nil
}

func appendHistoryToCpuUsageRankOfVm(rank []definition.VmPercentageUsage) {
	for i, vm := range rank {
		history, err := GetCpuHistoryOfVm(vm.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get cpu history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetCpuHistoryOfVm(entityId string, period definition.Period) ([]definition.TimeUsedPercent, error) {
	stmt := fmt.Sprintf(vmCpuUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageHistoryOfVm(c)
}

func parseCpuUsageHistoryOfVm(c *api.QueryTableResult) ([]definition.TimeUsedPercent, error) {
	points := []definition.TimeUsedPercent{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeUsedPercent{
				Time:        definition.TimeLocalISO8601(date),
				UsedPercent: parseUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetMemoryUsageRankOfVms(stmt string) ([]definition.VmMetricsUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseVmMemoryRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToMemoryUsageRankOfVm(rank)
	return rank, nil
}

func appendHistoryToMemoryUsageRankOfVm(rank []definition.VmMetricsUsage) {
	for i, vm := range rank {
		history, err := GetMemoryHistoryOfVm(vm.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get memory history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetMemoryHistoryOfVm(entityId string, period definition.Period) ([]definition.TimeUsedPercent, error) {
	stmt := fmt.Sprintf(vmMemoryUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageHistoryOfVm(c)
}

func parseMemoryUsageHistoryOfVm(c *api.QueryTableResult) ([]definition.TimeUsedPercent, error) {
	points := []definition.TimeUsedPercent{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeUsedPercent{
				Time:        definition.TimeLocalISO8601(date),
				UsedPercent: parseUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetDiskReadIopsRankOfVms(stmt string) ([]definition.VmDiskIopsUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parsDiskIopsRankOfVm(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToDiskReadIopsRankOfVm(rank)
	return rank, nil
}

func appendHistoryToDiskReadIopsRankOfVm(rank []definition.VmDiskIopsUsage) {
	for i, vm := range rank {
		history, err := GetDiskReadIopsHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get disk iops history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskReadIopsHistoryOfVm(entityId, device string) ([]definition.TimeOpsPoint, error) {
	stmt := fmt.Sprintf(vmStorageIopsReadHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func GetDiskWriteIopsRankOfVms(stmt string) ([]definition.VmDiskIopsUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parsDiskIopsRankOfVm(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToDiskWriteIopsRankOfVm(rank)
	return rank, nil
}

func appendHistoryToDiskWriteIopsRankOfVm(rank []definition.VmDiskIopsUsage) {
	for i, vm := range rank {
		history, err := GetDiskWriteIopsHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get disk iops history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskWriteIopsHistoryOfVm(entityId, device string) ([]definition.TimeOpsPoint, error) {
	stmt := fmt.Sprintf(vmStorageIopsWriteHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func GetNetworkTrafficInRankOfVms(stmt string) ([]definition.VmNetworkTrafficUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseVmNetworkPacketRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToNetworkTrafficInRankOfVm(rank)
	return rank, nil
}

func appendHistoryToNetworkTrafficInRankOfVm(rank []definition.VmNetworkTrafficUsage) {
	for i, vm := range rank {
		history, err := GetNetworkTrafficInHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get network traffic in history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficInHistoryOfVm(entityId, device string) ([]definition.TimePacketsPoint, error) {
	stmt := fmt.Sprintf(vmNetworkIngressHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func GetNetworkTrafficOutRankOfVms(stmt string) ([]definition.VmNetworkTrafficUsage, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseVmNetworkPacketRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToNetworkTrafficOutRankOfVm(rank)
	return rank, nil
}

func appendHistoryToNetworkTrafficOutRankOfVm(rank []definition.VmNetworkTrafficUsage) {
	for i, vm := range rank {
		history, err := GetNetworkTrafficOutHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get network traffic out history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficOutHistoryOfVm(entityId, device string) ([]definition.TimePacketsPoint, error) {
	stmt := fmt.Sprintf(vmNetworkEgressHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func parseCpuUsageRankOfHost(c *api.QueryTableResult) ([]definition.HostPercentageUsage, error) {
	rank := []definition.HostPercentageUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.HostPercentageUsage{
				Id:          parseHost(c.Record()),
				Name:        parseHost(c.Record()),
				UsedPercent: parseUsedOfHost(c.Record()),
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

func parseDiskLatencyHistory(c *api.QueryTableResult) ([]definition.TimeMillisecondPoint, error) {
	points := []definition.TimeMillisecondPoint{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().Local().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeMillisecondPoint{
				Time:        definition.TimeLocalISO8601(date),
				Millisecond: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parsDiskIopsRankOfVm(c *api.QueryTableResult) ([]definition.VmDiskIopsUsage, error) {
	rank := []definition.VmDiskIopsUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.VmDiskIopsUsage{
				Id:     parseResourceId(c.Record()),
				Name:   parseVmName(c.Record()),
				Device: parseDevice(c.Record()),
				Ops:    parseStorageUsedOfVm(c.Record()),
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

func genHostCpuUsageHistoryStmt(hostId string) string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range("start: -1h").
		Filter(fmt.Sprintf(`fn: (r) => r._measurement == "cpu" and r.host == "%s" and r._field == "usage_idle"`, hostId)).
		Map(`fn: (r) => ({ r with _value: 100.0 - r._value })`).
		Rename(`columns: {_value: "used"}`).
		String()
}

func parseVmNetworkPacketRank(c *api.QueryTableResult) ([]definition.VmNetworkTrafficUsage, error) {
	rank := []definition.VmNetworkTrafficUsage{}
	for c.Next() {
		rank = append(
			rank,
			definition.VmNetworkTrafficUsage{
				Id:      parseResourceId(c.Record()),
				Name:    parseVmName(c.Record()),
				Device:  parseDevice(c.Record()),
				Packets: parseUsed(c.Record()),
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

func parseUsedOfHost(record *query.FluxRecord) float64 {
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
