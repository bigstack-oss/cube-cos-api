package v1

const (
	Summary = "summary"
	Metrics = "metrics"
)

type DataCenterUsage struct {
	Cpu    ComputeStatistic `json:"cpu"`
	Memory SpaceStatistic   `json:"memory"`
}

type HostUsage struct {
	Cpu    ComputeStatistic `json:"cpu"`
	Memory SpaceStatistic   `json:"memory"`
}

type VmUsage struct {
	Vcpu    ComputeStatistic `json:"vcpu"`
	Memory  SpaceStatistic   `json:"memory"`
	Storage SpaceStatistic   `json:"storage"`
}

type ComputeStatistic struct {
	TotalCores  float64 `json:"totalCores"`
	UsedCores   float64 `json:"usedCores"`
	UsedPercent float64 `json:"usedPercent"`
	FreeCores   float64 `json:"freeCores"`
	FreePercent float64 `json:"freePercent"`
}

type SpaceStatistic struct {
	TotalMiB    float64 `json:"totalMiB"`
	UsedMiB     float64 `json:"usedMiB"`
	UsedPercent float64 `json:"usedPercent"`
	FreeMiB     float64 `json:"freeMiB"`
	FreePercent float64 `json:"freePercent"`
}

type TrafficStatistic struct {
	Ingress float64 `json:"ingress"`
	Egress  float64 `json:"egress"`
}

type MetricRank struct {
	Unit string      `json:"unit"`
	Rank []RankPoint `json:"rank"`
}

type MetricHistory struct {
	Unit    string      `json:"unit"`
	History []TimeValue `json:"history"`
}

type RankPoint struct {
	Id      string      `json:"id"`
	Name    string      `json:"name"`
	Device  string      `json:"device,omitempty"`
	Value   any         `json:"value"`
	History []TimeValue `json:"history"`
}

type HostPercentageUsage struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	UsedPercent float64           `json:"usedPercent"`
	History     []TimeUsedPercent `json:"history"`
}

type VmPercentageUsage struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	UsedPercent float64           `json:"usedPercent"`
	History     []TimeUsedPercent `json:"history"`
}

type VmMetricsUsage struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Device      string            `json:"device,omitempty"`
	UsedPercent float64           `json:"usedPercent"`
	History     []TimeUsedPercent `json:"history"`
}

type VmNetworkTrafficUsage struct {
	Id      string             `json:"id"`
	Name    string             `json:"name"`
	Device  string             `json:"device,omitempty"`
	Packets float64            `json:"packets"`
	History []TimePacketsPoint `json:"history"`
}

type VmDiskIopsUsage struct {
	Id      string         `json:"id"`
	Name    string         `json:"name"`
	Device  string         `json:"device,omitempty"`
	Ops     float64        `json:"ops"`
	History []TimeOpsPoint `json:"history"`
}

type HostNetworkPacket struct {
	Id      string             `json:"id"`
	Name    string             `json:"name"`
	Packets float64            `json:"packets"`
	History []TimePacketsPoint `json:"history"`
}

type StorageTimeSeries struct {
	Unit  string      `json:"unit"`
	Read  []TimeValue `json:"read"`
	Write []TimeValue `json:"write"`
}

type StorageIopsSeries struct {
	Unit  string         `json:"unit"`
	Read  []TimeOpsPoint `json:"read"`
	Write []TimeOpsPoint `json:"write"`
}

type StorageLatencySeries struct {
	Read  []TimeMillisecondPoint `json:"read"`
	Write []TimeMillisecondPoint `json:"write"`
}

type TimeBytesPoint struct {
	Time  string  `json:"time"`
	Bytes float64 `json:"bytes"`
}

type TimeOpsPoint struct {
	Time string  `json:"time"`
	Ops  float64 `json:"ops"`
}

type TimeMillisecondPoint struct {
	Time        string  `json:"time"`
	Millisecond float64 `json:"millisecond"`
}

type TimeUsedPercent struct {
	Time        string  `json:"time"`
	UsedPercent float64 `json:"usedPercent"`
}

type TimePacketsPoint struct {
	Time    string  `json:"time"`
	Packets float64 `json:"packets"`
}

type TimeValue struct {
	Time  string `json:"time"`
	Value any    `json:"value"`
}
