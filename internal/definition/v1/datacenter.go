package v1

const (
	DataCenters = "datacenters"
)

type DataCenter struct {
	Id          string `json:"id,omitempty" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Version     string `json:"version" bson:"version"`
	VirtualIp   string `json:"virtualIp" bson:"virtualIp"`
	IsLocal     bool   `json:"isLocal" bson:"isLocal"`
	IsHaEnabled bool   `json:"isHaEnabled" bson:"isHaEnabled"`
	UtcTimeZone string `json:"utcTimeZone,omitempty" bson:"utcTimeZone"`
	Additional  `json:"additional,omitempty" bson:"additional"`
}

type Additional struct {
	HelpUrl           string `json:"helpUrl,omitempty" bson:"helpUrl"`
	V1ApiDocUrl       string `json:"v1ApiDoc,omitempty" bson:"v1ApiDoc"`
	NodeLicenseStatus `json:"nodeLicenseStatus" bson:"nodeLicenseStatus"`
}

type NodeLicenseStatus struct {
	Valid     int `json:"valid" bson:"valid"`
	Expired   int `json:"expired" bson:"expired"`
	Unlicense int `json:"unlicense" bson:"unlicense"`
}
