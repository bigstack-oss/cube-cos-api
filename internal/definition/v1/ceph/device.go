package ceph

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	log "go-micro.dev/v5/logger"
)

type Device struct {
	Dev      string  `json:"dev"`
	Class    string  `json:"class"`
	Type     string  `json:"type"`
	Reweight float64 `json:"weight"`
	Osds     []Osd   `json:"daemons"`
}

type Osd struct {
	Id           string  `json:"id"`
	DeviceClass  string  `json:"device_class"`
	Reweight     float64 `json:"reweight,omitzero"`
	UsagePercent float64 `json:"usagePercent"`
	Status       string  `json:"status"`
}

type RawOsd struct {
	Nodes []Node `json:"nodes"`
}

type Node struct {
	Id          int     `json:"id"`
	DeviceClass string  `json:"device_class"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Reweight    float64 `json:"reweight"`
	Utilization float64 `json:"utilization"`
	Status      string  `json:"status"`
}

type RawDevice struct {
	DevId    string    `json:"devId"`
	Location []HostDev `json:"location"`
	Daemons  []string  `json:"daemons"`
}

type HostDev struct {
	Host string `json:"host"`
	Dev  string `json:"dev"`
	Path string `json:"path"`
}

func GetDeviceMapByHost(host string) (map[string]Device, error) {
	raws, err := ListRawDevicesByHost(host)
	if err != nil {
		return nil, err
	}

	devices, err := convertToDevices(raws)
	if err != nil {
		return nil, err
	}

	devMap := map[string]Device{}
	for _, device := range devices {
		devMap[device.Dev] = device
	}

	return devMap, nil
}

func ListRawDevicesByHost(host string) ([]RawDevice, error) {
	out, err := exec.Command("ceph", "device", "ls-by-host", host, "-f", "json").CombinedOutput()
	if err != nil {
		log.Errorf("ceph: failed to list devices by %s(%v)", host, err)
		return nil, err
	}

	devices := []RawDevice{}
	err = json.Unmarshal(out, &devices)
	if err != nil {
		log.Errorf("ceph: failed to unmarshal devices by %s(%v)", host, err)
		return nil, err
	}

	if len(devices) == 0 {
		log.Errorf("ceph: no devices found for host %s", host)
		return nil, nil
	}

	return devices, nil
}

func convertToDevices(raws []RawDevice) ([]Device, error) {
	devices := []Device{}
	for _, raw := range raws {
		for _, loc := range raw.Location {
			devices = append(devices, genDeviceByRaw(raw, loc))
			break
		}
	}

	for i, device := range devices {
		devices[i].Class = getOsdDeviceClass(device.Osds)
		devices[i].Reweight = getMaxOsdReweight(device.Osds)
	}

	return devices, nil
}

func getOsdDeviceClass(osds []Osd) string {
	if len(osds) == 0 {
		return ""
	}

	class := "unknown"
	for _, osd := range osds {
		return osd.DeviceClass
	}

	return class
}

func getMaxOsdReweight(osds []Osd) float64 {
	if len(osds) == 0 {
		return 0.0
	}

	maxReweight := osds[0].Reweight
	for _, osd := range osds {
		if osd.Reweight > maxReweight {
			maxReweight = osd.Reweight
		}
	}

	return maxReweight
}

func genDeviceByRaw(raw RawDevice, loc HostDev) Device {
	return Device{
		Dev:  loc.Dev,
		Osds: genOsdsByRaw(raw.Daemons),
	}
}

func genOsdsByRaw(daemons []string) []Osd {
	osds := []Osd{}
	for _, daemon := range daemons {
		if strings.Contains(daemon, "mon.") {
			continue
		}

		osd, err := GetOsdByDaemonId(daemon)
		if err == nil {
			osds = append(osds, *osd)
		}
	}

	return osds
}

func GetOsdByDaemonId(daemonId string) (*Osd, error) {
	out, err := exec.Command("ceph", "osd", "df", daemonId, "-f", "json").CombinedOutput()
	if err != nil {
		log.Errorf("ceph: failed to get osd by daemon id %s(%v)", daemonId, err)
		return nil, err
	}

	raw := RawOsd{}
	err = json.Unmarshal(out, &raw)
	if err != nil {
		log.Errorf("ceph: failed to unmarshal osd by daemon id %s(%v)", daemonId, err)
		return nil, err
	}

	if len(raw.Nodes) == 0 {
		err = errors.New("no osd found")
		log.Errorf("ceph: no osd found by daemon id %s(%v)", daemonId, err)
		return nil, err
	}

	osd := convertToOsd(raw.Nodes[0])
	return &osd, nil
}

func convertToOsd(raw Node) Osd {
	return Osd{
		Id:           raw.Name,
		DeviceClass:  raw.DeviceClass,
		Reweight:     raw.Reweight,
		UsagePercent: math.RoundDown(raw.Utilization, 4),
		Status:       raw.Status,
	}
}
