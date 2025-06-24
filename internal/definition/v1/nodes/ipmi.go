package nodes

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/network"
)

const (
	DefaultIpmiDeviceId = uint8(0)
	IpmiMarkerfile      = "/etc/appliance/state/ipmi_detected"
)

type Ipmi struct {
	Host     string `json:"host" bson:"host"`
	Ip       string `json:"ip" bson:"ip"`
	Port     int    `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type IpmiEnablement struct {
	IsSupported bool   `json:"isSupported" bson:"isSupported"`
	IsConnected bool   `json:"isConnected" bson:"isConnected"`
	Ip          string `json:"ip" bson:"ip"`
}

func (i *Ipmi) CheckInvalidValues() error {
	if !network.IsValidIPv4(i.Ip) {
		return fmt.Errorf("ipmi host ip should be provided")
	}

	if !network.IsValidPortRange(i.Port) {
		return fmt.Errorf("ipmi port should be between 1 and 65535")
	}

	if i.Username == "" {
		return fmt.Errorf("ipmi username should be provided")
	}

	if i.Password == "" {
		return fmt.Errorf("ipmi password should be provided")
	}

	return nil
}
