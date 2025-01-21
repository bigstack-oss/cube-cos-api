package v1

const (
	DataCenters = "datacenters"
)

type DataCenter struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	VirtualIp   string `json:"virtualIp"`
	IsLocal     bool   `json:"isLocal"`
	IsHaEnabled bool   `json:"isHaEnabled"`
}
