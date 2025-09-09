package storages

type Model struct {
	Vendor    string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Product   string `json:"product" yaml:"product" bson:"product"`
	Multipath `json:"multipath" yaml:"multipath" bson:"multipath"`
	Storage   Storage `json:"storage" yaml:"storage" bson:"storage"`
}

type Multipath struct {
	Defaults            []Conf `json:"defaults" yaml:"defaults" bson:"defaults"`
	Blacklist           `json:"blacklist" yaml:"blacklist" bson:"blacklist"`
	BlacklistExceptions Blacklist       `json:"blacklistExceptions" yaml:"blacklistExceptions" bson:"blacklistExceptions"`
	Devices             []ModelDevice   `json:"devices" yaml:"devices" bson:"devices"`
	Overrides           []Conf          `json:"overrides" yaml:"overrides" bson:"overrides"`
	Multipaths          []MultipathWwid `json:"multipaths" yaml:"multipaths" bson:"multipaths"`
}

type Blacklist struct {
	Devnode string        `json:"devnode" yaml:"devnode" bson:"devnode"`
	Devices []ModelDevice `json:"devices" yaml:"devices" bson:"devices"`
}

type ModelDevice struct {
	Vendor   string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Product  string `json:"product" yaml:"product" bson:"product"`
	Settings []Conf `json:"settings" yaml:"settings" bson:"settings"`
}

type MultipathWwid struct {
	WWID     string `json:"wwid" yaml:"wwid" bson:"wwid"`
	Settings []Conf `json:"settings" yaml:"settings" bson:"settings"`
}
