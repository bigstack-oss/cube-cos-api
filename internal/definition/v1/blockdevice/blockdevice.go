package blockdevice

import "fmt"

const (
	NetCode = "43"

	HDD = "HDD"
	SSD = "SSD"
)

type SmartCtl struct {
	State  string `json:"state" bson:"state"`
	Remark string `json:"remark" bson:"remark"`
}

func WithDevPath(name string) string {
	return fmt.Sprintf("/dev/%s", name)
}
