package metric

const (
	Module  = "metrics"
	Summary = "summary"

	TimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

type DataCenterUsage struct {
	Cpu    Compute `json:"cpu"`
	Memory Space   `json:"memory"`
}

type HostUsage struct {
	Cpu    Compute `json:"cpu"`
	Memory Space   `json:"memory"`
}

type VmUsage struct {
	Vcpu    Compute `json:"vcpu"`
	Memory  Space   `json:"memory"`
	Storage Space   `json:"storage"`
}

type Compute struct {
	TotalCores  float64 `json:"totalCores"`
	UsedCores   float64 `json:"usedCores"`
	UsedPercent float64 `json:"usedPercent"`
	FreeCores   float64 `json:"freeCores"`
	FreePercent float64 `json:"freePercent"`
}

type Space struct {
	TotalMiB    float64 `json:"totalMiB"`
	UsedMiB     float64 `json:"usedMiB"`
	UsedPercent float64 `json:"usedPercent"`
	FreeMiB     float64 `json:"freeMiB"`
	FreePercent float64 `json:"freePercent"`
}

type Traffic struct {
	Ingress float64 `json:"ingress"`
	Egress  float64 `json:"egress"`
}

type Rank struct {
	Unit string      `json:"unit"`
	Rank []RankPoint `json:"rank"`
}

type History struct {
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

type StorageTimeSeries struct {
	Unit  string      `json:"unit"`
	Read  []TimeValue `json:"read"`
	Write []TimeValue `json:"write"`
}

type TimeValue struct {
	Time  string `json:"time"`
	Value any    `json:"value"`
}
