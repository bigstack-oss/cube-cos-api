package blockdevice

const (
	NetCode = "43"

	HDD = "HDD"
	SSD = "SSD"
)

type SmartCtl struct {
	State  string `json:"state" bson:"state"`
	Remark string `json:"remark" bson:"remark"`
}
