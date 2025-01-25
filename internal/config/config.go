package config

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/log"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	yaml "github.com/go-micro/plugins/v5/config/encoder/yaml"
	jsonitor "github.com/json-iterator/go"
	"go-micro.dev/v5/config"
	"go-micro.dev/v5/config/reader"
	"go-micro.dev/v5/config/reader/json"
	"go-micro.dev/v5/config/source/file"
)

var (
	Opts Options
)

type Options struct {
	Kind     string `json:"kind" yaml:"kind"`
	Metadata `json:"metadata" yaml:"metadata"`
	Spec     `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name    string            `json:"name" yaml:"name"`
	Version string            `json:"version" yaml:"version"`
	Labels  map[string]string `json:"labels" yaml:"labels"`
}

type Spec struct {
	Listen          `json:"listen" yaml:"listen"`
	Store           `json:"store" yaml:"store"`
	Identity        `json:"identity" yaml:"identity"`
	ResourceControl `json:"resourceControl" yaml:"resourceControl"`
	Observability   `json:"observability" yaml:"observability"`
}

type Listen struct {
	Port    int `json:"port" yaml:"port"`
	Address `json:"Address" yaml:"address"`
}

type Address struct {
	Local     string `json:"local" yaml:"local"`
	Advertise string `json:"advertise" yaml:"advertise"`
}

type Identity struct {
	OsPolicy       string           `json:"osPolicy" yaml:"osPolicy"`
	LogoutRedirect string           `json:"logoutRedirect" yaml:"logoutRedirect"`
	Keycloak       keycloak.Options `json:"keycloak" yaml:"keycloak"`
	Saml           saml.Options     `json:"saml" yaml:"saml"`
}

type ResourceControl struct {
	Openstack openstack.Options `json:"openstack" yaml:"openstack"`
	K3s       `json:"k3s" yaml:"k3s"`
}

type K3s struct {
	Auth string `json:"auth" yaml:"auth"`
}

type Store struct {
	MongoDB  mongo.Options  `json:"mongodb" yaml:"mongodb"`
	InfluxDB influx.Options `json:"influxdb" yaml:"influxdb"`
}

type Observability struct {
	Log log.Options `json:"log" yaml:"log"`
}

func (o *Options) String() (string, error) {
	b, err := jsonitor.Marshal(o)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func InitServerOpts(filePath string) error {
	conf, err := newConfiger()
	if err != nil {
		return err
	}

	src := file.NewSource(file.WithPath(filePath))
	err = conf.Load(src)
	if err != nil {
		return err
	}

	err = conf.Get().Scan(&Opts)
	if err != nil {
		return err
	}

	return nil
}

func newConfiger() (config.Config, error) {
	return config.NewConfig(
		config.WithReader(
			json.NewReader(
				reader.WithEncoder(yaml.NewEncoder()),
			),
		),
	)
}
