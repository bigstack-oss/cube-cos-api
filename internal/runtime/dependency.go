package runtime

import (
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/aws"
	bshttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	bslog "github.com/bigstack-oss/bigstack-dependency-go/pkg/log"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func initDependencies() error {
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

	err = newGlobalAwsHelper(conf.Opts.Spec.Aws)
	if err != nil {
		log.Errorf("failed to init aws helper: %s", err.Error())
	}

	err = newGlobalKeycloakHelper(conf.Opts.Spec.Identity.Keycloak)
	if err != nil {
		log.Errorf("failed to init keycloak helper: %s", err.Error())
		return err
	}

	err = newKeycloakOidcAuth()
	if err != nil {
		log.Errorf("failed to init oidc auth in keycloak: %s", err.Error())
		return err
	}

	err = newDefaultOidcSecret()
	if err != nil {
		log.Errorf("failed to init oidc secret in keycloak: %s", err.Error())
		return err
	}

	err = newDefaultNodeToken()
	if err != nil {
		log.Errorf("failed to init node token: %s", err.Error())
		return err
	}

	err = newKeycloakSamlMapper()
	if err != nil {
		log.Errorf("failed to init saml mapper in keycloak: %s", err.Error())
		return err
	}

	err = newTuningSearchIndex()
	if err != nil {
		log.Errorf("failed to init tuning search index: %s", err.Error())
	}

	err = newTuningRecordTTL()
	if err != nil {
		log.Errorf("failed to init tuning record ttl: %s", err.Error())
	}

	err = newSupportFileSearchIndex()
	if err != nil {
		log.Errorf("failed to init support file search index: %s", err.Error())
	}

	err = newLicenseSearchIndex()
	if err != nil {
		log.Errorf("failed to init license search index: %s", err.Error())
	}

	err = newNodeSearchIndex()
	if err != nil {
		log.Errorf("failed to init node search index: %s", err.Error())
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
	if opts.Auth.Source == "file" {
		return openstack.NewGlobalHelper(
			openstack.AuthSource(opts.Auth.Source),
			openstack.AuthFile(opts.Auth.File),
			openstack.EnableAutoRenew(opts.Auth.EnableAutoRenew),
		)
	}

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

func newGlobalAwsHelper(opts aws.Options) error {
	return aws.NewGlobalHelper(
		aws.Region(opts.Region),
		aws.EnableStaticCreds(opts.EnableStaticCreds),
		aws.AccessKey(opts.AccessKey),
		aws.SecretKey(opts.SecretKey),
		aws.EnableCustomURL(opts.EnableCustomURL),
		aws.S3Url(v1.GetRadosGatewayUrl()),
		aws.InsecureSkipVerify(opts.InsecureSkipVerify),
	)
}

func newGlobalKeycloakHelper(opts keycloak.Options) error {
	if opts.Ip == "" {
		opts.Ip = v1.DataCenterVip
	}

	return keycloak.NewGlobalHelper(
		keycloak.Scheme(opts.Scheme),
		keycloak.Ip(opts.Ip),
		keycloak.Port(opts.Port),
		keycloak.Path(opts.Path),
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

func newKeycloakOidcAuth() error {
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
	if err != nil {
		log.Errorf("failed to login admin: %s", err.Error())
		return err
	}

	_, err = h.CreateClient(
		v1.DefaultKeycloakRealm,
		v1.DefaultOidcClientOpts,
	)
	if err == nil {
		return nil
	}
	if err.(*gocloak.APIError).Code == http.StatusConflict {
		return nil
	}

	return err
}

func newDefaultOidcSecret() error {
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
	if err != nil {
		log.Errorf("failed to login admin: %s", err.Error())
		return err
	}

	clients, err := h.GetClients(
		v1.DefaultKeycloakRealm,
		gocloak.GetClientsParams{ClientID: gocloak.StringP(v1.DefaultOidcClientId)},
	)
	if err != nil {
		log.Errorf("failed to get clients: %s", err.Error())
		return err
	}
	if len(clients) == 0 {
		return fmt.Errorf("oidc client not found")
	}

	secret, err := h.GetClientSecret(
		v1.DefaultKeycloakRealm,
		*clients[0].ID,
	)
	if err != nil {
		log.Errorf("failed to get client secret: %s", err.Error())
		return err
	}

	v1.DefaultOidcClientSecret = *secret.Value
	return nil
}

func newDefaultNodeToken() error {
	v1.DefaultNodeToken = v1.GenNodeToken(
		v1.Hostname,
		v1.AdvertiseAddr,
	)
	if v1.DefaultNodeToken == "" {
		return fmt.Errorf("failed to generate node token")
	}

	return nil
}

func newKeycloakSamlMapper() error {
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
	if err != nil {
		log.Errorf("failed to login admin: %s", err.Error())
		return err
	}

	samlId := saml.GenServiceProviderMetadataUrl(conf.Opts.Spec.Identity.Saml)
	clients, err := h.GetClients(
		v1.DefaultKeycloakRealm,
		gocloak.GetClientsParams{ClientID: gocloak.StringP(samlId.String())},
	)
	if err != nil {
		log.Errorf("failed to get clients: %s", err.Error())
		return err
	}
	if len(clients) == 0 {
		return fmt.Errorf("saml client not found")
	}

	_, err = h.CreateClientProtocolMapper(
		v1.DefaultKeycloakRealm,
		*clients[0].ID,
		genSamlMapper(),
	)
	if err == nil {
		return nil
	}
	if err.(*gocloak.APIError).Code == http.StatusConflict {
		return nil
	}

	return err
}

func genSamlMapper() gocloak.ProtocolMapperRepresentation {
	return gocloak.ProtocolMapperRepresentation{
		Name:           gocloak.StringP("username"),
		Protocol:       gocloak.StringP("saml"),
		ProtocolMapper: gocloak.StringP("saml-user-property-mapper"),
		Config: &map[string]string{
			"user.attribute":       "username",
			"attribute.name":       "username",
			"attribute.nameformat": "Basic",
		},
	}
}

func newTuningSearchIndex() error {
	return v1.InitTuningSearchIndex()
}

func newTuningRecordTTL() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.CreateExpirationIndex(
		v1.TuningDB(),
		v1.TuningReqCollection(),
		bson.D{{Key: "status.createdAt", Value: 1}},
		v1.TuningRecordTTL,
	)
}

func newSupportFileSearchIndex() error {
	return support.InitFileSearchIndex()
}

func newLicenseSearchIndex() error {
	return v1.InitLicenseSearchIndex()
}

func newNodeSearchIndex() error {
	return v1.InitNodeSearchIndex()
}
