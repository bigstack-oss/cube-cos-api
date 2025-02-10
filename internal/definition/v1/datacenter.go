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
	Additional  `json:"additional,omitempty" bson:"additional"`
}

type Additional struct {
	HelpUrl string `json:"helpUrl,omitempty" bson:"helpUrl"`
}

// M1 TODO: have to think about if we
// 1). need to add the Id of datacenter
// 2). if (1) is true, then what's factor to generate the Id
func (d *DataCenter) SetDetailsByInitedInfo() {
	d.Name = DataCenterName
	d.VirtualIp = DataCenterVip
	d.IsLocal = IsHaEnabled
	d.IsHaEnabled = IsHaEnabled
}
