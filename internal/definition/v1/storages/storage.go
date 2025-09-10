package storages

const (
	CinderConf = "/etc/cinder/cinder.conf"
)

type ReqOpts struct {
	Name string
}

type Cinder struct {
	Name       string `json:"name" yaml:"name" bson:"name"`
	IsExternal bool   `json:"isExternal" yaml:"isExternal" bson:"isExternal"`
	Device     `json:"device" yaml:"device" bson:"device"`
	Storage    `json:"storage" yaml:"storage" bson:"storage"`
}

type Device struct {
	Vendor  string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Product string `json:"product" yaml:"product" bson:"product"`
}

type Storage struct {
	Service    `json:"service" yaml:"service" bson:"service"`
	VolumeType `json:"volumeType" yaml:"volumeType" bson:"volumeType"`
	Image      `json:"image" yaml:"image" bson:"image"`
	UpdateTime string `json:"updateTime" yaml:"updateTime" bson:"updateTime"`
}

type Service struct {
	DriverSection []Conf         `json:"driverSection" yaml:"driverSection" bson:"driverSection"`
	ExtraSettings []ExtraSetting `json:"extraSettings" yaml:"extraSettings" bson:"extraSettings"`
}

type Conf struct {
	Key   string `json:"key" yaml:"key" bson:"key"`
	Value string `json:"value" yaml:"value" bson:"value"`
}

type ExtraSetting struct {
	SectionHeader string `json:"sectionHeader" yaml:"sectionHeader" bson:"sectionHeader"`
	Settings      []Conf `json:"settings" yaml:"settings" bson:"settings"`
}

type VolumeType struct {
	Settings []Conf `json:"settings" yaml:"settings" bson:"settings"`
}

type Image struct {
	UseMultipath   bool `json:"useMultipath" yaml:"useMultipath" bson:"useMultipath"`
	ForceMultipath bool `json:"forceMultipath" yaml:"forceMultipath" bson:"forceMultipath"`
}
