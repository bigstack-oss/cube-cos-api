package base

const (
	DataCenters = "datacenters"

	Cloud = "cloud"
	Edge  = "edge"
)

type DataCenter struct {
	Type        string   `json:"type" bson:"type"`
	Id          string   `json:"id,omitempty" bson:"id"`
	Name        string   `json:"name" bson:"name"`
	Roles       []string `json:"roles" bson:"roles"`
	Version     string   `json:"version" bson:"version"`
	VirtualIp   string   `json:"virtualIp" bson:"virtualIp"`
	IsLocal     bool     `json:"isLocal" bson:"isLocal"`
	IsHaEnabled bool     `json:"isHaEnabled" bson:"isHaEnabled"`
	UtcTimeZone string   `json:"utcTimeZone,omitempty" bson:"utcTimeZone"`
	Additional  `json:"additional" bson:"additional"`
}

type Additional struct {
	HelpUrl           string `json:"helpUrl,omitempty" bson:"helpUrl"`
	V1ApiDocUrl       string `json:"v1ApiDocUrl,omitempty" bson:"v1ApiDocUrl"`
	NodeLicenseStatus `json:"nodeLicenseStatus" bson:"nodeLicenseStatus"`
}

type NodeLicenseStatus struct {
	Valid     int `json:"valid" bson:"valid"`
	Expired   int `json:"expired" bson:"expired"`
	Unlicense int `json:"unlicense" bson:"unlicense"`
}
