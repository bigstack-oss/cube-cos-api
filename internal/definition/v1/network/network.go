package network

import "net"

func IsValidIPv4(ip string) bool {
	parsedIp := net.ParseIP(ip)
	return parsedIp != nil && parsedIp.To4() != nil
}

func IsValidPortRange(port int) bool {
	return port > 0 && port <= 65535
}
