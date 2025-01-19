package config

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/log"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	yaml "github.com/go-micro/plugins/v5/config/encoder/yaml"
	"go-micro.dev/v5/config"
	"go-micro.dev/v5/config/reader"
	"go-micro.dev/v5/config/reader/json"
	"go-micro.dev/v5/config/source/file"
)

var (
	Data Payload
)

type Payload struct {
	Kind     string `json:"kind" yaml:"kind"`
	Metadata `json:"metadata" yaml:"metadata"`
	Spec     `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name string `json:"name" yaml:"name"`
}

type Spec struct {
	Runtime    string `json:"runtime" yaml:"runtime"`
	Listen     `json:"listen" yaml:"listen"`
	Store      `json:"store" yaml:"store"`
	Auth       `json:"auth" yaml:"auth"`
	Dependency `json:"dependency" yaml:"dependency"`
	Log        log.Options `json:"log" yaml:"log"`
}

type Dependency struct {
	CubeCos   string            `json:"cubeCos" yaml:"cubeCos"`
	Openstack openstack.Options `json:"openstack" yaml:"openstack"`
	K3s       string            `json:"k3s" yaml:"k3s"`
}

type Listen struct {
	Port    int `json:"port" yaml:"port"`
	Address `json:"Address" yaml:"address"`
}

type Store struct {
	MongoDB  mongo.Options  `json:"mongodb" yaml:"mongodb"`
	InfluxDB influx.Options `json:"influxdb" yaml:"influxdb"`
}

type Auth struct {
	Keycloak keycloak.Options `json:"keycloak" yaml:"keycloak"`
}

type Address struct {
	Local     string `json:"local" yaml:"local"`
	Advertise string `json:"advertise" yaml:"advertise"`
}

func NewConfiger() (config.Config, error) {
	return config.NewConfig(
		config.WithReader(
			json.NewReader(
				reader.WithEncoder(yaml.NewEncoder()),
			),
		),
	)
}

func Load(filePath string) (config.Config, error) {
	configer, err := NewConfiger()
	if err != nil {
		return nil, err
	}

	confSrc := file.NewSource(file.WithPath(filePath))
	err = configer.Load(confSrc)
	if err != nil {
		return nil, err
	}

	return configer, nil
}
