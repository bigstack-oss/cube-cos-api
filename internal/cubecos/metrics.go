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
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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

func IsMetricTypeValid(t string) bool {
	return isMetricTypeValid[t]
}

func IsEntityTypeValid(t string) bool {
	return isEntityTypeValid[t]
}

func IsMetricReportTypeValid(t string) bool {
	return isMetricReportTypeValid[t]
}

func SyncDataCenterMetricsSummary() {
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

func GetMetricsSummary() *Summary {
	return metricsSummary
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
		UsedCores:   math.RoundDown(usedCores, 4),
		UsedPercent: math.RoundDown(usedPercent/float64(len(cpuStats)), 4),
		FreeCores:   math.RoundDown(freeCores, 4),
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
		TotalMiB:    math.RoundDown(totalMiB, 4),
		UsedMiB:     math.RoundDown(usedMiB, 4),
		UsedPercent: math.RoundDown(usedPercent/float64(len(spaceStats)), 4),
		FreeMiB:     math.RoundDown(freeMiB, 4),
		FreePercent: math.RoundDown(freePercent/float64(len(spaceStats)), 4),
	}
}

func GetHostSummary() (*HostSummary, error) {
	s := &HostSummary{}
	nodes := definition.ListNodes()
	s.SetHostUsageByNodes(nodes)
	s.SetRoleUsageByHosts()
	return s, nil
}

func GetHostUsage(node definition.Node) (*definition.HostUsage, error) {
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
		return askPeerHostForCpuSummary(hostname)
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

func GetCpuHistoryOfHost(stmt string) ([]definition.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageHistory(c)
}

func parseCpuUsageHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: parseUsedOfHost(c.Record()),
			},
		)
	}

	return points, nil
}

func askPeerHostForCpuSummary(hostname string) (*definition.ComputeStatistic, error) {
	node, err := definition.GetNodeByHostname(hostname)
	if err != nil {
		return nil, err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&cubeapi.ComputeStatisticData{}).
		SetHeader(node.GenAuthHeader()).
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

	computeStatistic := resp.Result().(*cubeapi.ComputeStatisticData).Data
	return &computeStatistic, nil
}

func GetCpuUsageRankOfHosts(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToCpuUsageRank(rank []definition.RankPoint) {
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
	usages := []definition.SpaceStatistic{}
	for _, node := range definition.ListNodes() {
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
		SetResult(&cubeapi.SpaceStatisticData{}).
		SetHeader(node.GenAuthHeader()).
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

	spaceStatistic := resp.Result().(*cubeapi.SpaceStatisticData).Data
	return &spaceStatistic, nil
}

func GetMemoryUsageRankOfHosts(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToMemoryUsageRank(rank []definition.RankPoint) {
	for i, host := range rank {
		stmt := fmt.Sprintf(hostMemoryUsageHistoryStmt, host.Id)
		history, err := GetMemoryHistoryOfHost(stmt)
		if err != nil {
			log.Errorf("failed to get memory history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetMemoryHistoryOfHost(stmt string) ([]definition.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageHistory(c)
}

func parseMemoryUsageHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: parseUsedOfHost(c.Record()),
			},
		)
	}

	return points, nil
}

func GetDiskStorageSummaryOfHost() (*definition.SpaceStatistic, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	stmt := genDiskStorageSummaryOfHostStmt()
	influx := influx.GetGlobalHelper()
	c, err := influx.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer c.Close()
	spaceStatistic := &definition.SpaceStatistic{}
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

func GetDiskStorageBandwidthHistory(readStmt, writeStmt string) (*definition.StorageTimeSeries, error) {
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

	return &definition.StorageTimeSeries{
		Unit:  "bytes",
		Read:  read,
		Write: write,
	}, nil
}

func GetDiskIopsHistoryOfHosts(readStmt, writeStmt string) (*definition.StorageTimeSeries, error) {
	readSeries, err := getDiskIopsHistoryOfHosts(readStmt)
	if err != nil {
		return nil, err
	}

	writeSeries, err := getDiskIopsHistoryOfHosts(writeStmt)
	if err != nil {
		return nil, err
	}

	return &definition.StorageTimeSeries{
		Unit:  "ops",
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func GeDiskLatencyHistoryOfHosts(readStmt, writeStmt string) (*definition.StorageTimeSeries, error) {
	readSeries, err := geDiskLatencyHistoryOfHosts(readStmt)
	if err != nil {
		return nil, err
	}

	writeSeries, err := geDiskLatencyHistoryOfHosts(writeStmt)
	if err != nil {
		return nil, err
	}

	return &definition.StorageTimeSeries{
		Unit:  "milliseconds",
		Read:  readSeries,
		Write: writeSeries,
	}, nil
}

func getDiskIopsHistoryOfHosts(stmt string) ([]definition.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func geDiskLatencyHistoryOfHosts(stmt string) ([]definition.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskLatencyHistory(c)
}

func GetDiskUsageRankOfHosts(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToDiskUsageRank(rank []definition.RankPoint) {
	for i, host := range rank {
		history, err := GetDiskUsageHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get disk usage history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskUsageHistoryOfHost(entityId string, period definition.Period) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(hostDiskUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskUsageHistory(c)
}

func parseDiskUsageHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: parseUsedOfHost(c.Record()),
			},
		)
	}

	return points, nil
}

func GetNetworkTrafficInRankOfHosts() (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func appendHistoryToNetworkTrafficInRank(rank []definition.RankPoint) {
	for i, host := range rank {
		history, err := GetNetworkTrafficInHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get network traffic in history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficInHistoryOfHost(entityId string, period definition.Period) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(hostNetworkIngressHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func parseNetworkTrafficHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func GetNetworkTrafficOutRankOfHosts() (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func appendHistoryToNetworkTrafficOutRank(rank []definition.RankPoint) {
	for i, host := range rank {
		history, err := GetNetworkTrafficOutHistoryOfHost(host.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get network traffic out history of host %s: %v", host.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficOutHistoryOfHost(entityId string, period definition.Period) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(hostNetworkEgressHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func GetCpuUsageRankOfVms(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToCpuUsageRankOfVm(rank []definition.RankPoint) {
	for i, vm := range rank {
		history, err := GetCpuHistoryOfVm(vm.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get cpu history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetCpuHistoryOfVm(entityId string, period definition.Period) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(vmCpuUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseCpuUsageHistoryOfVm(c)
}

func parseCpuUsageHistoryOfVm(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: parseUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetMemoryUsageRankOfVms(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "percentage",
		Rank: rank,
	}, nil
}

func appendHistoryToMemoryUsageRankOfVm(rank []definition.RankPoint) {
	for i, vm := range rank {
		history, err := GetMemoryHistoryOfVm(vm.Id, definition.Period{})
		if err != nil {
			log.Errorf("failed to get memory history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetMemoryHistoryOfVm(entityId string, period definition.Period) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(vmMemoryUsageHistoryStmt, entityId)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseMemoryUsageHistoryOfVm(c)
}

func parseMemoryUsageHistoryOfVm(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: parseUsed(c.Record()),
			},
		)
	}

	return points, nil
}

func GetDiskReadIopsRankOfVms(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "ops",
		Rank: rank,
	}, nil
}

func appendHistoryToDiskReadIopsRankOfVm(rank []definition.RankPoint) {
	for i, vm := range rank {
		history, err := GetDiskReadIopsHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get disk iops history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskReadIopsHistoryOfVm(entityId, device string) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(vmStorageIopsReadHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func GetDiskWriteIopsRankOfVms(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "ops",
		Rank: rank,
	}, nil
}

func appendHistoryToDiskWriteIopsRankOfVm(rank []definition.RankPoint) {
	for i, vm := range rank {
		history, err := GetDiskWriteIopsHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get disk iops history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetDiskWriteIopsHistoryOfVm(entityId, device string) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(vmStorageIopsWriteHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskOpsHistory(c)
}

func GetNetworkTrafficInRankOfVms(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func appendHistoryToNetworkTrafficInRankOfVm(rank []definition.RankPoint) {
	for i, vm := range rank {
		history, err := GetNetworkTrafficInHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get network traffic in history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func GetNetworkTrafficInHistoryOfVm(entityId, device string) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(vmNetworkIngressHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func GetNetworkTrafficOutRankOfVms(stmt string) (*definition.MetricRank, error) {
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
	return &definition.MetricRank{
		Unit: "packets",
		Rank: rank,
	}, nil
}

func GetNetworkTrafficOutHistoryOfVm(entityId, device string) ([]definition.TimeValue, error) {
	stmt := fmt.Sprintf(vmNetworkEgressHistoryStmt, entityId, device)
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseNetworkTrafficHistory(c)
}

func genDiskStorageSummaryOfHostStmt() string {
	query := influx.Query{}
	return query.Bucket("telegraf").
		Range(`start: -1h`).
		Filter(`fn: (r) => r._measurement == "disk"`).
		Filter(fmt.Sprintf(`fn: (r) => r.host == "%s"`, definition.Hostname)).
		Sort(`columns: ["_time"], desc: true`).
		Pivot(`rowKey: ["_time","host"], columnKey: ["_field"], valueColumn: "_value"`).
		Limit(`n: 1`).
		String()
}

func appendHistoryToNetworkTrafficOutRankOfVm(rank []definition.RankPoint) {
	for i, vm := range rank {
		history, err := GetNetworkTrafficOutHistoryOfVm(vm.Id, vm.Device)
		if err != nil {
			log.Errorf("failed to get network traffic out history of vm %s: %v", vm.Id, err)
			continue
		}

		rank[i].History = history
	}
}

func parseCpuUsageRankOfHost(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
				Id:    parseHost(c.Record()),
				Name:  parseHost(c.Record()),
				Value: parseUsedOfHost(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func parseVmCpuUsageRank(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
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

func parseMemoryUsageRankOfHost(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
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

func parseVmMemoryRank(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
				Id:    parseResourceId(c.Record()),
				Name:  parseVmName(c.Record()),
				Value: parseCpuUsedOfVm(c.Record()),
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

func parseHostStorageUsageRank(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
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

func parseDiskOpsHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parseDiskLatencyHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func parsDiskIopsRankOfVm(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
				Id:     parseResourceId(c.Record()),
				Name:   parseVmName(c.Record()),
				Device: parseDevice(c.Record()),
				Value:  parseStorageUsedOfVm(c.Record()),
			},
		)
	}
	if c.Err() != nil {
		return nil, c.Err()
	}

	return rank, nil
}

func getDiskBandwidthHistory(stmt string) ([]definition.TimeValue, error) {
	c, cancel, err := influx.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer cancel()
	defer c.Close()
	return parseDiskBandwidthHistory(c)
}

func parseDiskBandwidthHistory(c *api.QueryTableResult) ([]definition.TimeValue, error) {
	points := []definition.TimeValue{}
	for c.Next() {
		date, err := time.Parse(eventTimeLayout, c.Record().Time().String())
		if err != nil {
			continue
		}

		points = append(
			points,
			definition.TimeValue{
				Time:  definition.TimeLocalRFC3339(date),
				Value: math.RoundDown(c.Record().Value().(float64), 4),
			},
		)
	}

	return points, nil
}

func hostNetworkIngressRank(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
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

func parseVmNetworkPacketRank(c *api.QueryTableResult) ([]definition.RankPoint, error) {
	rank := []definition.RankPoint{}
	for c.Next() {
		rank = append(
			rank,
			definition.RankPoint{
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
