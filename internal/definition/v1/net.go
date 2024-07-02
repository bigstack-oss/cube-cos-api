package v1

import (
	"fmt"
	"net"
)

const (
	NetMajorInterface = "eth0"
)

func GetMacAddr(interfaceName string) (string, error) {
	nets, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	macAddr := ""
	for _, net := range nets {
		if net.Name == interfaceName {
			macAddr = net.HardwareAddr.String()
			break
		}
	}
	if macAddr == "" {
		return "", fmt.Errorf("mac address not found from interface: %s", interfaceName)
	}

	return macAddr, nil
}
