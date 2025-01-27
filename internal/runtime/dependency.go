package runtime

import (
	bshttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	bslog "github.com/bigstack-oss/bigstack-dependency-go/pkg/log"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	log "go-micro.dev/v5/logger"
)

func initDependencyHelpers() error {
	err := newGlobalLogHelper(conf.Opts.Spec.Observability.Log)
	if err != nil {
		log.Errorf("failed to init logger: %s", err.Error())
		return err
	}

	err = newGlobalHttpHelper()
	if err != nil {
		log.Errorf("failed to init http helper: %s", err.Error())
		return err
	}

	err = newGlobalSamlHelper()
	if err != nil {
		log.Errorf("failed to init keycloak auth: %s", err.Error())
		return err
	}

	err = newGlobalMongoHelper(conf.Opts.Spec.Store.MongoDB)
	if err != nil {
		log.Errorf("failed to init mongo helper: %s", err.Error())
		return err
	}

	err = newGlobalInfluxHelper(conf.Opts.Spec.Store.InfluxDB)
	if err != nil {
		log.Errorf("failed to init influx helper: %s", err.Error())
		return err
	}

	err = newGlobalOpenstackHelper(conf.Opts.Spec.Openstack)
	if err != nil {
		log.Errorf("failed to init openstack helper: %s", err.Error())
		return err
	}

	err = newGlobalKeycloakHelper(conf.Opts.Spec.Identity.Keycloak)
	if err != nil {
		log.Errorf("failed to init keycloak helper: %s", err.Error())
		return err
	}

	return nil
}

func newGlobalLogHelper(opts bslog.Options) error {
	return bslog.NewGlobalHelper(
		bslog.File(opts.File),
		bslog.Level(opts.Level),
		bslog.Backups(opts.Rotation.Backups),
		bslog.Size(opts.Rotation.Size),
		bslog.TTL(opts.Rotation.TTL),
		bslog.Compress(opts.Rotation.Compress),
	)
}

func newGlobalMongoHelper(opts mongo.Options) error {
	return mongo.NewGlobalHelper(
		mongo.Uri(opts.Uri),
		mongo.AuthEnable(opts.Auth.Enable),
		mongo.AuthSource(opts.Auth.Source),
		mongo.AuthUsername(opts.Auth.Username),
		mongo.AuthPassword(opts.Auth.Password),
		mongo.ReplicaSet(opts.ReplicaSet),
	)
}

func newGlobalInfluxHelper(opts influx.Options) error {
	return influx.NewGlobalHelper(
		influx.Url(opts.Url),
	)
}
func newGlobalOpenstackHelper(opts openstack.Options) error {
	return openstack.NewGlobalHelper(
		openstack.AuthType(opts.Auth.Type),
		openstack.AuthUrl(opts.Auth.Url),
		openstack.EnableAutoRenew(opts.Auth.EnableAutoRenew),
		openstack.ProjectName(opts.Auth.Project.Name),
		openstack.ProjectDomainName(opts.Auth.Project.Domain.Name),
		openstack.Username(opts.Auth.Username),
		openstack.Password(opts.Auth.Password),
	)
}

func newGlobalKeycloakHelper(opts keycloak.Options) error {
	return keycloak.NewGlobalHelper(
		keycloak.Host(opts.Host),
		keycloak.Realm(opts.Realm),
		keycloak.Username(opts.Username),
		keycloak.Password(opts.Password),
		keycloak.Insecure(opts.TlsInsecureSkipVerify),
	)
}

func newGlobalHttpHelper() error {
	return bshttp.NewGlobalHelper()
}

func newGlobalSamlHelper() error {
	return saml.NewGlobalAuth(conf.Opts.Spec.Identity.Saml)
}
