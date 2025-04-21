package v1

const (
	NetBlockDeviceCode = "43"
)

type SmartCtl struct {
	SmartStatus `json:"smart_status" bson:"smart_status"`
}

type SmartStatus struct {
	Passed bool `json:"passed" bson:"passed"`
}
