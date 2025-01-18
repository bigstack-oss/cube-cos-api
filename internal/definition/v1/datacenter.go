package v1

const (
	DataCenters = "datacenters"
)

type DataCenter struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	VirtualIp   string `json:"virtualIp"`
	IsLocal     bool   `json:"isLocal"`
	IsHaEnabled bool   `json:"isHaEnabled"`
}
