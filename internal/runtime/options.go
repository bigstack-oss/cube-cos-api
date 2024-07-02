package runtime

import "github.com/bigstack-oss/bigstack-dependency-go/pkg/log"

var (
	conf Config
)

type Config struct {
	Kind     string `json:"kind"`
	Metadata `json:"metadata"`
	Spec     `json:"spec"`
}

type Metadata struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
}

type Spec struct {
	Runtime string `json:"runtime"`
	Auth    `json:"auth"`
	Listen  `json:"listen"`
	Log     log.Options `json:"log"`
}

type Auth struct {
	Openstack string `json:"openstack"`
	K3s       string `json:"k3s"`
}

type Listen struct {
	Port    int `json:"port"`
	Address `json:"Address"`
}

type Address struct {
	Local     string `json:"local"`
	Advertise string `json:"advertise"`
}
