package cubecos

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	cubeapi "github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auth"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	log "go-micro.dev/v5/logger"
)

var (
	summaryUpdate  = sync.Mutex{}
	metricsSummary *Summary

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

func IsValidMetricType(t string) bool {
	return isMetricTypeValid[t]
}

func IsEntityTypeValid(t string) bool {
	return isEntityTypeValid[t]
}

func IsMetricReportTypeValid(t string) bool {
	return isMetricReportTypeValid[t]
}

func SyncMetricsSummary() {
	summaryUpdate.Lock()
	defer summaryUpdate.Unlock()

	summary, err := syncDataCenterSummary()
	if err != nil {
		log.Errorf("metrics: failed to sync data center summary: %v", err)
		return
	}

	metricsSummary = summary
}

func syncDataCenterSummary() (*Summary, error) {
	host, err := GetHostSummary()
	if err != nil {
		log.Errorf("metrics: failed to get host summary: %v", err)
		return nil, err
	}

	dataCenter, err := GetDataCenterUsage(host)
	if err != nil {
		log.Errorf("metrics: failed to get data center usage: %v", err)
		return nil, err
	}

	vm, err := GetVmSummary()
	if err != nil {
		log.Errorf("metrics: failed to get vm summary: %v", err)
		return nil, err
	}

	return &Summary{
		DataCenter: *dataCenter,
		Host:       *host,
		Vm:         *vm,
	}, nil
}

func GetMetricsSummary() *Summary {
	return metricsSummary
}

func GetDataCenterUsage(hostSummary *HostSummary) (*DataCenterSummary, error) {
	return &DataCenterSummary{
		Usage: metric.DataCenterUsage{
			Cpu:    GetHostsCpuAverage(hostSummary.ListCpuUsages()),
			Memory: GetHostsMemoryAverage(hostSummary.ListMemoryUsages()),
		},
	}, nil
}

func GetHostsCpuAverage(cpuStats []metric.Compute) metric.Compute {
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

	return metric.Compute{
		TotalCores:  totalCores,
		UsedCores:   math.RoundDown(usedCores, 4),
		UsedPercent: math.RoundDown(usedPercent/float64(len(cpuStats)), 4),
		FreeCores:   math.RoundDown(freeCores, 4),
		FreePercent: math.RoundDown(freePercent/float64(len(cpuStats)), 4),
	}
}

func GetHostsMemoryAverage(spaceStats []metric.Space) metric.Space {
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

	return metric.Space{
		TotalMiB:    math.RoundDown(totalMiB, 4),
		UsedMiB:     math.RoundDown(usedMiB, 4),
		UsedPercent: math.RoundDown(usedPercent/float64(len(spaceStats)), 4),
		FreeMiB:     math.RoundDown(freeMiB, 4),
		FreePercent: math.RoundDown(freePercent/float64(len(spaceStats)), 4),
	}
}

func GetHostSummary() (*HostSummary, error) {
	s := &HostSummary{}
	nodes := nodes.List()
	s.SetHostUsageByNodes(nodes)
	s.SetRoleUsageByHosts()
	return s, nil
}

func GetHostUsage(node nodes.Node) (*metric.HostUsage, error) {
	if node.IsDown() {
		return nil, fmt.Errorf("host %s is down", node.Hostname)
	}

	cpuStat, err := GetHostCpuSummary(node.Hostname)
	if err != nil {
		log.Errorf("metrics: failed to get cpu summary of host %s: %v", node.Hostname, err)
		return nil, err
	}

	memoryStat, err := GetHostMemoryUsageSummary(node.Hostname)
	if err != nil {
		log.Errorf("metrics: failed to get memory summary of host %s: %v", node.Hostname, err)
		return nil, err
	}

	return &metric.HostUsage{
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

func GetHostsCpuSummary(stmt string) (*metric.Compute, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageOfHost(c)
}

func GetHostCpuSummary(hostname string) (*metric.Compute, error) {
	if !nodes.IsLocal(hostname) {
		return askPeerNodeCpuSummary(hostname)
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

	return &metric.Compute{
		TotalCores:  float64(runtime.NumCPU()),
		UsedCores:   math.RoundDown(usedCores, 4),
		UsedPercent: math.RoundDown(cpuPercents[0], 4),
		FreeCores:   math.RoundDown(freeCores, 4),
		FreePercent: math.RoundDown(100-cpuPercents[0], 4),
	}, nil
}

func GetHostCpuHistory(stmt string) (*metric.History, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	history, err := parseCpuUsageHistory(c)
	if err != nil {
		log.Errorf("metrics: failed to parse cpu usage history: %v", err)
		return nil, err
	}

	return &metric.History{
		Unit:    "percentage",
		History: history,
	}, nil
}

func parseCpuUsageHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: parseHostUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func askPeerNodeCpuSummary(hostname string) (*metric.Compute, error) {
	node, err := nodes.Get(hostname)
	if err != nil {
		return nil, err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&cubeapi.ComputeStatistic{}).
		SetHeaders(auth.GetNodeSecret()).
		Get(node.GetMetricUrl("cpuUsage", "summary"))
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return &resp.Result().(*cubeapi.ComputeStatistic).Data, nil
	}

	return nil, fmt.Errorf(
		"failed to get cpu usage of host %s: %d %s",
		hostname,
		resp.StatusCode(),
		string(resp.Body()),
	)
}

func GetHostsCpuUsageRank(stmt string) (*metric.Rank, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parseHostCpuUsageRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToCpuUsageRank(rank)
	return &metric.Rank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToCpuUsageRank(rank []metric.RankPoint) {
	for i, host := range rank {
		data, err := GetHostCpuHistory(genHostCpuUsageHistoryStmt(host.Id))
		if err != nil {
			log.Errorf("metrics: failed to get cpu history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = data.History
	}
}

func GetHostsMemoryUsageSummary() (*metric.Space, error) {
	usages := []metric.Space{}
	for _, node := range nodes.List() {
		usage, err := GetHostMemoryUsageSummary(node.Hostname)
		if err != nil {
			continue
		}

		usages = append(usages, *usage)
	}

	stat := GetHostsMemoryAverage(usages)
	return &stat, nil
}

func GetHostMemoryUsageSummary(hostname string) (*metric.Space, error) {
	if !nodes.IsLocal(hostname) {
		return askPeerNodeForMemorySummary(hostname)
	}

	stat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	usedPercent := stat.UsedPercent
	freePercent := 100.0 - usedPercent
	return &metric.Space{
		TotalMiB:    math.RoundDown(float64(stat.Total)/1024/1024, 4),
		UsedMiB:     math.RoundDown(float64(stat.Used)/1024/1024, 4),
		FreeMiB:     math.RoundDown(float64(stat.Total-stat.Used)/1024/1024, 4),
		UsedPercent: math.RoundDown(usedPercent, 4),
		FreePercent: math.RoundDown(freePercent, 4),
	}, nil
}

func askPeerNodeForMemorySummary(hostname string) (*metric.Space, error) {
	node, err := nodes.Get(hostname)
	if err != nil {
		return nil, err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&cubeapi.SpaceStatistic{}).
		SetHeaders(auth.GetNodeSecret()).
		Get(node.GetMetricUrl("memoryUsage", "summary"))
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return &resp.Result().(*cubeapi.SpaceStatistic).Data, nil
	}

	return nil, fmt.Errorf(
		"failed to get memory usage of host %s: %d %s",
		hostname,
		resp.StatusCode(),
		string(resp.Body()),
	)
}

func GetHostsMemoryUsageRank(stmt string) (*metric.Rank, error) {
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
	return &metric.Rank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToMemoryUsageRank(rank []metric.RankPoint) {
	for i, host := range rank {
		stmt := fmt.Sprintf(hostMemoryUsageHistoryStmt, host.Id)
		history, err := GetHostMemoryHistory(stmt)
		if err != nil {
			log.Errorf("metrics: failed to get memory history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetHostMemoryHistory(stmt string) ([]metric.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageHistory(c)
}

func GetHostMemorySizeHistory(stmt string) (*metric.History, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	history, err := parseMemorySizeHistory(c)
	if err != nil {
		log.Errorf("metrics: failed to parse memory usage history: %v", err)
		return nil, err
	}

	return &metric.History{
		Unit:    "sizeMiB",
		History: history,
	}, nil
}

func parseMemoryUsageHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: parseHostUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func parseMemorySizeHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		value := float64(parseSizeOfHost(c.Record())) / 1024.0 / 1024.0
		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: math.RoundDown(value, 4),
			},
		)
	}

	return points, nil
}

func GetHostDiskStorageSummary() (*metric.Space, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	stmt := genHostDiskStorageSummaryStmt()
	influx := influx.GetGlobalHelper()
	c, err := influx.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer c.Close()
	spaceStatistic := &metric.Space{}
	for c.Next() {
		record := c.Record()
		spaceStatistic.TotalMiB = parseDiskTotal(record)
		spaceStatistic.UsedMiB = parseDiskUsed(record)
		spaceStatistic.FreeMiB = parseDiskFree(record)
		spaceStatistic.UsedPercent = parseDiskUsedPercent(record)
		spaceStatistic.FreePercent = math.RoundDown(100-spaceStatistic.UsedPercent, 4)
	}

	return spaceStatistic, nil
}

func GetHostsDiskBandwidthHistory(readStmt, writeStmt string) (*metric.StorageTimeSeries, error) {
	read, err := getDiskBandwidthHistory(readStmt)
	if err != nil {
		log.Errorf("metrics: failed to get host storage read bandwidth series: %v", err)
		return nil, err
	}

	write, err := getDiskBandwidthHistory(writeStmt)
	if err != nil {
		log.Errorf("metrics: failed to get host storage write bandwidth series: %v", err)
		return nil, err
	}

	return &metric.StorageTimeSeries{
		Unit:  "bytes",
		Read:  read,
		Write: write,
	}, nil
}

func GetHostsDiskIopsHistory(readStmt, writeStmt string) (*metric.StorageTimeSeries, error) {
	readSeries, err := getHostsDiskIopsHistory(readStmt)
	if err != nil {
		return nil, err
	}

	writeSeries, err := getHostsDiskIopsHistory(writeStmt)
	if err != nil {
		return nil, err
	}

	return &metric.StorageTimeSeries{
		Unit:  "ops",
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func GeHostsDiskLatencyHistory(readStmt, writeStmt string) (*metric.StorageTimeSeries, error) {
	readSeries, err := getHostsDiskLatencyHistory(readStmt)
	if err != nil {
		return nil, err
	}

	writeSeries, err := getHostsDiskLatencyHistory(writeStmt)
	if err != nil {
		return nil, err
	}

	return &metric.StorageTimeSeries{
		Unit:  "milliseconds",
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func getHostsDiskIopsHistory(stmt string) ([]metric.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func getHostsDiskLatencyHistory(stmt string) ([]metric.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskLatencyHistory(c)
}

func GetHostsDiskUsageRank(stmt string) (*metric.Rank, error) {
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
	return &metric.Rank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToDiskUsageRank(rank []metric.RankPoint) {
	for i, host := range rank {
		history, err := GetHostDiskUsageHistory(host.Id, v1.Period{})
		if err != nil {
			log.Errorf("metrics: failed to get disk usage history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetHostDiskUsageHistory(entityId string, period v1.Period) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(hostDiskUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskUsageHistory(c)
}

func parseDiskUsageHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: parseHostUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetHostsNetworkIngressRank() (*metric.Rank, error) {
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
	return &metric.Rank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func appendHistoryToNetworkTrafficInRank(rank []metric.RankPoint) {
	for i, host := range rank {
		history, err := GetHostNetworkIngressHistory(host.Id, v1.Period{})
		if err != nil {
			log.Errorf("metrics: failed to get network traffic in history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetHostNetworkIngressHistory(entityId string, period v1.Period) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(hostNetworkIngressHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func parseNetworkTrafficHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func GetHostsNetworkEgressRank() (*metric.Rank, error) {
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

	appendHistoryToNetworkEgressRank(rank)
	return &metric.Rank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func appendHistoryToNetworkEgressRank(rank []metric.RankPoint) {
	for i, host := range rank {
		history, err := GetHostNetworkEgressHistory(host.Id, v1.Period{})
		if err != nil {
			log.Errorf("metrics: failed to get network traffic out history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetHostNetworkEgressHistory(entityId string, period v1.Period) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(hostNetworkEgressHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func GetVmsCpuUsageRank(stmt string) (*metric.Rank, error) {
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

	appendHistoryToVmCpuUsageRank(rank)
	return &metric.Rank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToVmCpuUsageRank(rank []metric.RankPoint) {
	for i, vm := range rank {
		history, err := GetVmCpuHistory(vm.Id, v1.Period{})
		if err != nil {
			log.Errorf("metrics: failed to get cpu history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetVmCpuHistory(entityId string, period v1.Period) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(vmCpuUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseVmCpuUsageHistory(c)
}

func parseVmCpuUsageHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: parseUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetVmsMemoryUsageRank(stmt string) (*metric.Rank, error) {
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

	appendHistoryToVmMemoryUsageRank(rank)
	return &metric.Rank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToVmMemoryUsageRank(rank []metric.RankPoint) {
	for i, vm := range rank {
		history, err := GetMemoryHistoryOfVm(vm.Id, v1.Period{})
		if err != nil {
			log.Errorf("metrics: failed to get memory history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetMemoryHistoryOfVm(entityId string, period v1.Period) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(vmMemoryUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageHistoryOfVm(c)
}

func parseMemoryUsageHistoryOfVm(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: parseUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetVmsDiskReadIopsRank(stmt string) (*metric.Rank, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parsVmDiskIopsRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToVmDiskReadIopsRank(rank)
	return &metric.Rank{
		Unit: "ops",
		Rank: rank,
	}, nil
}

func appendHistoryToVmDiskReadIopsRank(rank []metric.RankPoint) {
	for i, vm := range rank {
		history, err := GetVmDiskReadIopsHistory(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("metrics: failed to get disk iops history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetVmDiskReadIopsHistory(entityId, device string) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(vmStorageIopsReadHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func GetVmsDiskWriteIopsRank(stmt string) (*metric.Rank, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	rank, err := parsVmDiskIopsRank(c)
	if err != nil {
		return nil, err
	}

	appendHistoryToDiskWriteIopsRankOfVm(rank)
	return &metric.Rank{
		Unit: "ops",
		Rank: rank,
	}, nil
}

func appendHistoryToDiskWriteIopsRankOfVm(rank []metric.RankPoint) {
	for i, vm := range rank {
		history, err := GetDiskWriteIopsHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("metrics: failed to get disk iops history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskWriteIopsHistoryOfVm(entityId, device string) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(vmStorageIopsWriteHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func GetVmsNetworkIngressRank(stmt string) (*metric.Rank, error) {
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

	appendHistoryToVmNetworkIngressRank(rank)
	return &metric.Rank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func appendHistoryToVmNetworkIngressRank(rank []metric.RankPoint) {
	for i, vm := range rank {
		history, err := GetVmNetworkTrafficInHistory(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("metrics: failed to get network traffic in history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetVmNetworkTrafficInHistory(entityId, device string) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(vmNetworkIngressHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func GetVmsNetworkEgressRank(stmt string) (*metric.Rank, error) {
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

	appendHistoryToVmNetworkEgressRank(rank)
	return &metric.Rank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func GetVmNetworkTrafficOutHistory(entityId, device string) ([]metric.TimeValue, error) {
	stmt := fmt.Sprintf(vmNetworkEgressHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func genHostDiskStorageSummaryStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range(`start: -1h`).
		Filter(`fn: (r) => r._measurement == "disk"`).
		Filter(fmt.Sprintf(`fn: (r) => r.host == "%s"`, base.Hostname)).
		Sort(`columns: ["_time"], desc: true`).
		Pivot(`rowKey: ["_time","host"], columnKey: ["_field"], valueColumn: "_value"`).
		Limit(`n: 1`).
		String()
}

func appendHistoryToVmNetworkEgressRank(rank []metric.RankPoint) {
	for i, vm := range rank {
		history, err := GetVmNetworkTrafficOutHistory(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("metrics: failed to get network traffic out history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func parseHostCpuUsageRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:    parseHost(c.Record()),
				Name:  parseHost(c.Record()),
				Value: parseHostUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseVmCpuUsageRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:    parseResourceId(c.Record()),
				Name:  parseVmName(c.Record()),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseCpuUsageOfHost(c *api.QueryTableResult) (*metric.Compute, error) {
	usage := metric.Compute{}
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

func parseCpuUsage(record *query.FluxRecord) metric.Compute {
	usedPercent := record.Value().(float64)
	return metric.Compute{
		UsedPercent: math.RoundDown(usedPercent, 4),
		FreePercent: math.RoundDown(100-usedPercent, 4),
	}
}

func parseMemoryUsageRankOfHost(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:    parseHost(c.Record()),
				Name:  parseHost(c.Record()),
				Value: parseUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseVmMemoryRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:    parseResourceId(c.Record()),
				Name:  parseVmName(c.Record()),
				Value: parseVmCpuUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseMemoryUsageSummaryOfHost(c *api.QueryTableResult) (*metric.Space, error) {
	memoryUsage := metric.Space{}
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

func parseMemoryUsage(record *query.FluxRecord) metric.Space {
	usedPercent := record.Value().(float64)
	return metric.Space{
		UsedPercent: math.RoundDown(usedPercent, 4),
		FreePercent: math.RoundDown(100-usedPercent, 4),
	}
}

func parseHostStorageUsageRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:    parseHost(c.Record()),
				Name:  parseHost(c.Record()),
				Value: parseUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseDiskOpsHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parseDiskLatencyHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parsVmDiskIopsRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:     parseResourceId(c.Record()),
				Name:   parseVmName(c.Record()),
				Device: parseDevice(c.Record()),
				Value:  parseVmStorageUsed(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func getDiskBandwidthHistory(stmt string) ([]metric.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskBandwidthHistory(c)
}

func parseDiskBandwidthHistory(c *api.QueryTableResult) ([]metric.TimeValue, error) {
	points := []metric.TimeValue{}
	for c.Next() {
		date, err := time.Parse(event.TimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			metric.TimeValue{
				Time:  v1.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func hostNetworkIngressRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:    parseHost(c.Record()),
				Name:  parseHost(c.Record()),
				Value: parseUsed(c.Record()),
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

func parseVmNetworkPacketRank(c *api.QueryTableResult) ([]metric.RankPoint, error) {
	rank := []metric.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			metric.RankPoint{
				Id:     parseResourceId(c.Record()),
				Name:   parseVmName(c.Record()),
				Device: parseDevice(c.Record()),
				Value:  parseUsed(c.Record()),
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
		return ""
	}

	return host
}

func parseHostUsed(record *query.FluxRecord) float64 {
	used, ok := record.ValueByKey("used").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseSizeOfHost(record *query.FluxRecord) float64 {
	used, ok := record.ValueByKey("used").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseVmCpuUsed(record *query.FluxRecord) float64 {
	used, ok := record.ValueByKey("used").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(used, 4)
}

func parseVmStorageUsed(record *query.FluxRecord) float64 {
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

func parseDiskTotal(record *query.FluxRecord) float64 {
	total, ok := record.ValueByKey("total").(int64)
	if !ok {
		return 0
	}

	return math.RoundDown(float64(total)/1024.0/1024.0, 4)
}

func parseDiskUsed(record *query.FluxRecord) float64 {
	total, ok := record.ValueByKey("used").(int64)
	if !ok {
		return 0
	}

	return math.RoundDown(float64(total)/1024.0/1024.0, 4)
}

func parseDiskFree(record *query.FluxRecord) float64 {
	total, ok := record.ValueByKey("free").(int64)
	if !ok {
		return 0
	}

	return math.RoundDown(float64(total)/1024.0/1024.0, 4)
}

func parseDiskUsedPercent(record *query.FluxRecord) float64 {
	usedPercent, ok := record.ValueByKey("used_percent").(float64)
	if !ok {
		return 0
	}

	return math.RoundDown(usedPercent, 4)
}
