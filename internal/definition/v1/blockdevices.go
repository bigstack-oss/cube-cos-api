package v1

const (
	NetBlockDeviceCode = "43"
)

type BlockDeviceStatus struct {
	State  string `json:"state" bson:"state"`
	Remark string `json:"remark" bson:"remark"`
}
