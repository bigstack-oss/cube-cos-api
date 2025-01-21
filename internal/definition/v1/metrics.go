package v1

const (
	Summary = "summary"
)

type Metrics struct {
	Vcpu    ComputeStatistic `json:"vcpu"`
	Memory  SpaceStatistic   `json:"memory"`
	Storage SpaceStatistic   `json:"storage"`
}

type ComputeStatistic struct {
	TotalCores int `json:"totalCores"`
	UsedCores  int `json:"usedCores"`
	FreeCores  int `json:"freeCores"`
}

type SpaceStatistic struct {
	TotalMiB float64 `json:"totalMiB"`
	UsedMiB  float64 `json:"usedMiB"`
	FreeMiB  float64 `json:"freeMiB"`
}
