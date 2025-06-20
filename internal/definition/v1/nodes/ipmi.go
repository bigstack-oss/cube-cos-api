package nodes

const (
	DefaultIpmiDeviceId = uint8(0)
)

type Ipmi struct {
	Enabled  bool `json:"enabled" bson:"enabled"`
	Host     `json:"host" bson:"host"`
	Port     int    `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}
