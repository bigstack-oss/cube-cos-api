package base

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"runtime"

	log "go-micro.dev/v5/logger"
)

const (
	NetMajorInterface = "eth0"

	VarRunCubeCosApiDir = "/var/run/cube-cos-api"
	DataCenterHelpUrl   = "https://www.bigstack.co/contact-us"
	BoardSerialPath     = "/sys/class/dmi/id/board_serial"
)

var (
	SystemSeed               string
	ServiceDiscoveryIdentity string
	BoardSerial              string
	DataCenterName           string
	ActiveFirmwareVersion    string
	ActiveFirmwareUpdatedAt  string
	InactiveFirmwareVersion  string
	FixpackVersion           string
	FixpackUpdatedAt         string
	DataCenterNumericVersion string
	DataCenterVip            string
	HostID                   string
	Hostname                 string
	SerialNumber             string
	CurrentRole              string
	ListenIp                 string
	ListenAddr               string
	ListenPort               int
	AdvertiseIp              string
	AdvertiseAddr            string
	AdvertisePort            int
	ManagementNet            string
	ManagementIp             string
	StorageNet               string
	StorageIP                string
	IsDomainNameEnabled      bool
	IsHaEnabled              bool
	IsGpuEnabled             bool
	NodeMetadata             map[string]string
)

func GetMacAddr(netInterface string) (string, error) {
	nets, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	macAddr := ""
	for _, net := range nets {
		if net.Name == netInterface {
			macAddr = net.HardwareAddr.String()
			break
		}
	}
	if macAddr == "" {
		return "", fmt.Errorf("mac address not found from interface: %s", netInterface)
	}

	return macAddr, nil
}

func CapturePanic() {
	recovery := recover()
	if recovery != nil {
		buf := make([]byte, 1<<16)
		stackSize := runtime.Stack(buf, true)
		log.Errorf("panic: captured %v\n stack trace:\n%s", recovery, buf[:stackSize])
	}
}

func GenerateNodeHashByMacAddr() (string, error) {
	macAddr, err := GetMacAddr(NetMajorInterface)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(macAddr))
	return hex.EncodeToString(hash[:])[:8], nil
}

func IsSuccessCode(err error) bool {
	if err == nil {
		return true
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return result.ExitCode() == 0
}

func GenApiDocUrl() string {
	return fmt.Sprintf(
		"https://%s/api/v1/datacenters/%s/apidocs/index.html",
		DataCenterVip,
		DataCenterName,
	)
}
