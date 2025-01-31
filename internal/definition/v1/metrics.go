package v1

const (
	Summary = "summary"
	Metrics = "metrics"
)

type Usage struct {
	Vcpu    ComputeStatistic `json:"vcpu"`
	Memory  SpaceStatistic   `json:"memory"`
	Storage SpaceStatistic   `json:"storage"`
}

type ComputeStatistic struct {
	TotalCores  int     `json:"totalCores"`
	UsedCores   int     `json:"usedCores"`
	UsedPercent float64 `json:"usedPercent"`
	FreeCores   int     `json:"freeCores"`
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

type HostPercentageUsage struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	UsedPercent float64 `json:"usedPercent"`
	FreePercent float64 `json:"freePercent"`
}

type VmPercentageUsage struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	UsedPercent float64 `json:"usedPercent"`
	FreePercent float64 `json:"freePercent"`
}

type VmMetricsUsage struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Device      string  `json:"device,omitempty"`
	Usage       float64 `json:"usage"`
	UsedPercent float64 `json:"usedPercent"`
	FreePercent float64 `json:"freePercent"`
}

type HostNetworkPacket struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Packets float64 `json:"packets"`
}

type StorageBandwidthSeries struct {
	Read  []TimeBytesPoint `json:"read"`
	Write []TimeBytesPoint `json:"write"`
}

type StorageIopsSeries struct {
	Read  []TimeOpsPoint `json:"read"`
	Write []TimeOpsPoint `json:"write"`
}

type StorageLatencySeries struct {
	Read  []TimeLatencyPoint `json:"read"`
	Write []TimeLatencyPoint `json:"write"`
}

type TimeBytesPoint struct {
	Time  string  `json:"time"`
	Bytes float64 `json:"bytes"`
}

type TimeOpsPoint struct {
	Time string  `json:"time"`
	Ops  float64 `json:"ops"`
}

type TimeLatencyPoint struct {
	Time string  `json:"time"`
	Ms   float64 `json:"ms"`
}
