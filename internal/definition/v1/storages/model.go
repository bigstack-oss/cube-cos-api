package storages

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

type Model struct {
	Driver    string `json:"driver" yaml:"driver" bson:"driver"`
	Vendor    string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Type      string `json:"type" yaml:"type" bson:"type"`
	Multipath []Path `json:"multipath" yaml:"multipath" bson:"multipath"`
	Storage   `json:"storage" yaml:"storage" bson:"storage"`
	Status    status.Model `json:"status" yaml:"status" bson:"status"`
}

type Path struct {
	Section     string       `json:"section" yaml:"section" bson:"section"`
	Attributes  []Attribute  `json:"attributes" yaml:"attributes" bson:"attributes"`
	SubSections []Subsection `json:"subsections" yaml:"subsections" bson:"subsections"`
}

type Subsection struct {
	Section    string      `json:"section" yaml:"section" bson:"section"`
	Attributes []Attribute `json:"attributes" yaml:"attributes" bson:"attributes"`
}

type Blacklist struct {
	Devnode string        `json:"devnode" yaml:"devnode" bson:"devnode"`
	Devices []ModelDevice `json:"devices" yaml:"devices" bson:"devices"`
}

type ModelDevice struct {
	Vendor   string      `json:"vendor" yaml:"vendor" bson:"vendor"`
	Product  string      `json:"product" yaml:"product" bson:"product"`
	Settings []Attribute `json:"settings" yaml:"settings" bson:"settings"`
}

type MultipathWwid struct {
	WWID     string      `json:"wwid" yaml:"wwid" bson:"wwid"`
	Settings []Attribute `json:"settings" yaml:"settings" bson:"settings"`
}
