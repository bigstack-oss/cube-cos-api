package v1

const (
	NetBlockDeviceCode = "43"
)

type SmartCtl struct {
	PassStatus `json:"passed" bson:"passed"`
}

type PassStatus struct {
	Passed bool `json:"passed" bson:"passed"`
}
