package base

const (
	DataCenters = "datacenters"

	Cloud = "cloud"
	Edge  = "edge"
)

type DataCenter struct {
	Type        string   `json:"type"`
	Id          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Roles       []string `json:"roles"`
	Version     string   `json:"version"`
	VirtualIp   string   `json:"virtualIp"`
	IsLocal     bool     `json:"isLocal"`
	IsHaEnabled bool     `json:"isHaEnabled"`
	UtcTimeZone string   `json:"utcTimeZone,omitempty"`
	Firmware    `json:"firmware"`
	Fixpack     `json:"fixpack"`
	Additional  `json:"additional"`
}

type Additional struct {
	HelpUrl           string `json:"helpUrl,omitempty"`
	V1ApiDocUrl       string `json:"v1ApiDocUrl,omitempty"`
	NodeLicenseStatus `json:"nodeLicenseStatus"`
}

type NodeLicenseStatus struct {
	Valid     int `json:"valid"`
	Expired   int `json:"expired"`
	Unlicense int `json:"unlicense"`
}

type Firmware struct {
	Version   string `json:"version"`
	UpdatedAt string `json:"updatedAt"`
}

type Fixpack struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	UpdatedAt string `json:"updatedAt"`
}
