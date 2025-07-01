package nodes

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

type NetworkInterface struct {
	Interface   string `json:"interface" yaml:"interface" bson:"interface"`
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busIdSlaves" yaml:"busIdSlaves" bson:"busIdSlaves"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

type RawNetworkInterface struct {
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busid" yaml:"busid" bson:"busid"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

type BlockDevice struct {
	Serial       string             `json:"serial" bson:"serial"`
	Name         string             `json:"device" yaml:"device" bson:"device"`
	Type         string             `json:"type" yaml:"type" bson:"type"`
	SizeMiB      float64            `json:"sizeMiB" yaml:"sizeMiB" bson:"sizeMiB"`
	Availability string             `json:"availability" yaml:"availability" bson:"availability"`
	Status       status.BlockDevice `json:"status" yaml:"status" bson:"status"`
}

// note:
// rota is named by lsblk tool, it means rotational device like HDD
type RawBlockDevice struct {
	Type        string   `json:"type"`
	Serial      string   `json:"serial"`
	Name        string   `json:"name"`
	Size        string   `json:"size"`
	Rota        bool     `json:"rota"`
	MountPoints []string `json:"mountpoints"`
}

type ImpiValidation struct {
	Board   `json:"board"`
	Product `json:"product"`
}

type Board struct {
	ManufacturingDate string `json:"manufacturingDate"`
	Manufacturer      string `json:"manufacturer"`
	Product           string `json:"product"`
	Serial            string `json:"serial"`
	PartNumber        string `json:"partNumber"`
}

type Product struct {
	Manufacturer string `json:"manufacturer"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Serial       string `json:"serial"`
}

func (r *RawBlockDevice) IsPartition() bool {
	return r.Type == "part"
}

func (r *RawBlockDevice) IsBlock() bool {
	return r.Type == "disk"
}

func (r *RawBlockDevice) NoMountPoints() bool {
	return len(r.MountPoints) == 0
}
