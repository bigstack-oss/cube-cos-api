package config

import (
	"flag"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/aws"
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
	confile string
	Opts    Options
)

func init() {
	flag.StringVar(&confile, "conf", "/etc/cube-cos-api/cube-cos-api.yaml", "")
	flag.StringVar(&Opts.Spec.Listen.Address.Local, "listen.address.local", Opts.Spec.Listen.Address.Local, "")
	flag.StringVar(&Opts.Spec.Listen.Address.Advertise, "listen.address.advertise", Opts.Spec.Listen.Address.Advertise, "")
	flag.IntVar(&Opts.Spec.Listen.Port, "listen.port", Opts.Spec.Listen.Port, "")
	flag.StringVar(&Opts.Spec.Identity.Os.Serial, "identity.os.serial", Opts.Spec.Identity.Os.Serial, "/sys/class/dmi/id/product_serial")
	flag.StringVar(&Opts.Spec.Identity.Os.Policy, "identity.os.policy", Opts.Spec.Identity.Os.Policy, "")
	flag.StringVar(&Opts.Spec.Identity.Os.System, "identity.os.system", Opts.Spec.Identity.Os.System, "")
	flag.StringVar(&Opts.Spec.Identity.Os.Hostname, "identity.os.hostname", Opts.Spec.Identity.Os.Hostname, "")
	flag.StringVar(&Opts.Spec.Identity.LogoutRedirect, "identity.logoutRedirect", Opts.Spec.Identity.LogoutRedirect, "")
	flag.StringVar(&Opts.Spec.Identity.Keycloak.Host.Scheme, "identity.keycloak.host.scheme", Opts.Spec.Identity.Keycloak.Host.Scheme, "")
	flag.StringVar(&Opts.Spec.Identity.Keycloak.Ip, "identity.keycloak.host.ip", Opts.Spec.Identity.Keycloak.Ip, "")
	flag.IntVar(&Opts.Spec.Identity.Keycloak.Port, "identity.keycloak.host.port", Opts.Spec.Identity.Keycloak.Port, "")
	flag.StringVar(&Opts.Spec.Identity.Keycloak.Path, "identity.keycloak.host.path", Opts.Spec.Identity.Keycloak.Path, "")
	flag.BoolVar(&Opts.Spec.Identity.Keycloak.TlsInsecureSkipVerify, "identity.keycloak.host.tlsInsecureSkipVerify", Opts.Spec.Identity.Keycloak.TlsInsecureSkipVerify, "")
	flag.StringVar(&Opts.Spec.Identity.Keycloak.Realm, "identity.keycloak.realm", Opts.Spec.Identity.Keycloak.Realm, "")
	flag.StringVar(&Opts.Spec.Identity.Keycloak.Auth.Username, "identity.keycloak.auth.username", Opts.Spec.Identity.Keycloak.Auth.Username, "")
	flag.StringVar(&Opts.Spec.Identity.Keycloak.Auth.Password, "identity.keycloak.auth.password", Opts.Spec.Identity.Keycloak.Auth.Password, "")
	flag.BoolVar(&Opts.Spec.Identity.Keycloak.TlsInsecureSkipVerify, "identity.keycloak.tlsInsecureSkipVerify", Opts.Spec.Identity.Keycloak.TlsInsecureSkipVerify, "")
	flag.StringVar(&Opts.Spec.Identity.Saml.IdentityProvider.MetadataPath, "identity.saml.identityProvider.metadataPath", Opts.Spec.Identity.Saml.IdentityProvider.MetadataPath, "")
	flag.StringVar(&Opts.Spec.Identity.Saml.IdentityProvider.Host.Scheme, "identity.saml.identityProvider.host.scheme", Opts.Spec.Identity.Saml.IdentityProvider.Host.Scheme, "")
	flag.IntVar(&Opts.Spec.Identity.Saml.IdentityProvider.Host.Port, "identity.saml.identityProvider.host.port", Opts.Spec.Identity.Saml.IdentityProvider.Host.Port, "")
	flag.BoolVar(&Opts.Spec.Identity.Saml.IdentityProvider.TlsInsecureSkipVerify, "identity.saml.identityProvider.tlsInsecureSkipVerify", Opts.Spec.Identity.Saml.IdentityProvider.TlsInsecureSkipVerify, "")
	flag.StringVar(&Opts.Spec.Identity.Saml.ServiceProvider.MetadataPath, "identity.saml.serviceProvider.metadataPath", Opts.Spec.Identity.Saml.ServiceProvider.MetadataPath, "")
	flag.StringVar(&Opts.Spec.Identity.Saml.ServiceProvider.Host.Scheme, "identity.saml.serviceProvider.host.scheme", Opts.Spec.Identity.Saml.ServiceProvider.Host.Scheme, "")
	flag.IntVar(&Opts.Spec.Identity.Saml.ServiceProvider.Host.Port, "identity.saml.serviceProvider.host.port", Opts.Spec.Identity.Saml.ServiceProvider.Host.Port, "")
	flag.BoolVar(&Opts.Spec.Identity.Saml.ServiceProvider.TlsInsecureSkipVerify, "identity.saml.serviceProvider.tlsInsecureSkipVerify", Opts.Spec.Identity.Saml.ServiceProvider.TlsInsecureSkipVerify, "")
	flag.StringVar(&Opts.Spec.ResourceControl.K3s.Auth, "resourceControl.k3s.auth", Opts.Spec.ResourceControl.K3s.Auth, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.ConfFile, "resourceControl.openstack.confFile", Opts.Spec.ResourceControl.Openstack.ConfFile, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Source, "resourceControl.openstack.auth.source", Opts.Spec.ResourceControl.Openstack.Auth.Source, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.File, "resourceControl.openstack.auth.file", Opts.Spec.ResourceControl.Openstack.Auth.File, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Type, "resourceControl.openstack.auth.type", Opts.Spec.ResourceControl.Openstack.Auth.Type, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Url, "resourceControl.openstack.auth.url", Opts.Spec.ResourceControl.Openstack.Auth.Url, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Username, "resourceControl.openstack.auth.username", Opts.Spec.ResourceControl.Openstack.Auth.Username, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Password, "resourceControl.openstack.auth.password", Opts.Spec.ResourceControl.Openstack.Auth.Password, "")
	flag.BoolVar(&Opts.Spec.ResourceControl.Openstack.Auth.EnableAutoRenew, "resourceControl.openstack.auth.enableAutoRenew", Opts.Spec.ResourceControl.Openstack.Auth.EnableAutoRenew, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Project.Name, "resourceControl.openstack.auth.project.name", Opts.Spec.ResourceControl.Openstack.Auth.Project.Name, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Openstack.Auth.Project.Domain.Name, "resourceControl.openstack.auth.project.domain.name", Opts.Spec.ResourceControl.Openstack.Auth.Project.Domain.Name, "")
	flag.StringVar(&Opts.Spec.ResourceControl.Aws.AccessKey, "resourceControl.aws.accessKey", Opts.Spec.ResourceControl.Aws.AccessKey, "admin")
	flag.StringVar(&Opts.Spec.ResourceControl.Aws.SecretKey, "resourceControl.aws.secretKey", Opts.Spec.ResourceControl.Aws.SecretKey, "admin")
	flag.StringVar(&Opts.Spec.Store.InfluxDB.Url, "store.influxdb.url", Opts.Spec.Store.InfluxDB.Url, "")
	flag.StringVar(&Opts.Spec.Store.MongoDB.Uri, "store.mongodb.uri", Opts.Spec.Store.MongoDB.Uri, "")
	flag.StringVar(&Opts.Spec.Store.MongoDB.Database, "store.mongodb.database", Opts.Spec.Store.MongoDB.Database, "")
	flag.StringVar(&Opts.Spec.Store.MongoDB.ReplicaSet, "store.mongodb.replicaSet", Opts.Spec.Store.MongoDB.ReplicaSet, "")
	flag.BoolVar(&Opts.Spec.Store.MongoDB.Auth.Enable, "store.mongodb.auth.enable", Opts.Spec.Store.MongoDB.Auth.Enable, "")
	flag.StringVar(&Opts.Spec.Store.MongoDB.Auth.Username, "store.mongodb.auth.username", Opts.Spec.Store.MongoDB.Auth.Username, "")
	flag.StringVar(&Opts.Spec.Store.MongoDB.Auth.Password, "store.mongodb.auth.password", Opts.Spec.Store.MongoDB.Auth.Password, "")
	flag.IntVar(&Opts.Spec.Observability.Log.Level, "observability.log.level", Opts.Spec.Observability.Log.Level, "")
	flag.StringVar(&Opts.Spec.Observability.Log.File, "observability.log.file", Opts.Spec.Observability.Log.File, "")
	flag.IntVar(&Opts.Spec.Observability.Log.Rotation.Backups, "observability.log.rotation.backups", Opts.Spec.Observability.Log.Rotation.Backups, "")
	flag.IntVar(&Opts.Spec.Observability.Log.Rotation.Size, "observability.log.rotation.size", Opts.Spec.Observability.Log.Rotation.Size, "")
	flag.IntVar(&Opts.Spec.Observability.Log.Rotation.TTL, "observability.log.rotation.ttl", Opts.Spec.Observability.Log.Rotation.TTL, "")
	flag.BoolVar(&Opts.Spec.Observability.Log.Rotation.Compress, "observability.log.rotation.compress", Opts.Spec.Observability.Log.Rotation.Compress, "")
	flag.Parse()
}

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
	Os             `json:"os" yaml:"os"`
	LogoutRedirect string           `json:"logoutRedirect" yaml:"logoutRedirect"`
	Keycloak       keycloak.Options `json:"keycloak" yaml:"keycloak"`
	Saml           saml.Options     `json:"saml" yaml:"saml"`
}

type Os struct {
	Serial   string `json:"serial" yaml:"serial"`
	Policy   string `json:"policy" yaml:"policy"`
	System   string `json:"system" yaml:"system"`
	Hostname string `json:"hostname" yaml:"hostname"`
}

type ResourceControl struct {
	Openstack openstack.Options `json:"openstack" yaml:"openstack"`
	K3s       `json:"k3s" yaml:"k3s"`
	Aws       aws.Options `json:"aws" yaml:"aws"`
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

func SyncOptions() error {
	err := parseFileOpts()
	if err != nil {
		return err
	}

	overrideOptsByFlags()
	return nil
}

func parseFileOpts() error {
	conf, err := newConfiger()
	if err != nil {
		return err
	}

	src := file.NewSource(file.WithPath(confile))
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

func overrideOptsByFlags() {
	flag.Parse()
}
