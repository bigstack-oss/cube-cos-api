package cubecos

type Summary struct {
	Vm      `json:"vm"`
	Role    `json:"role"`
	Metrics `json:"metrics"`
}

type Vm struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Suspend int `json:"suspend"`
	Paused  int `json:"paused"`
	Error   int `json:"error"`
	Unknown int `json:"unknown"`
}

type Role struct {
	ControlConverged int `json:"controlConverged"`
	Control          int `json:"control"`
	Compute          int `json:"compute"`
	Storage          int `json:"storage"`
	Others           int `json:"others"`
}

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
