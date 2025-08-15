package nodes

type Firmware struct {
	Active   string `json:"active,omitempty" yaml:"active,omitempty" bson:"active,omitempty"`
	Inactive string `json:"inactive,omitempty" yaml:"inactive,omitempty" bson:"inactive,omitempty"`
}
