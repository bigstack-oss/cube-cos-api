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
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func initDependencies() error {
	err := newGlobalHelpers()
	if err != nil {
		return err
	}

	err = newAuthIdentities()
	if err != nil {
		return err
	}

	err = newSearchIndexers()
	if err != nil {
		return err
	}

	err = newReqPendingTTLs()
	if err != nil {
		return err
	}

	return nil
}

func newGlobalHelpers() error {
	err := newGlobalLogHelper()
	if err != nil {
		log.Errorf("runtime: failed to init logger: %s", err.Error())
		return err
	}

	err = newGlobalHttpHelper()
	if err != nil {
		log.Errorf("runtime: failed to init http helper: %s", err.Error())
		return err
	}

	err = newGlobalSamlHelper()
	if err != nil {
		log.Errorf("runtime: failed to init keycloak auth: %s", err.Error())
		return err
	}

	err = newGlobalMongoHelper()
	if err != nil {
		log.Errorf("runtime: failed to init mongo helper: %s", err.Error())
		return err
	}

	err = newGlobalInfluxHelper()
	if err != nil {
		log.Errorf("runtime: failed to init influx helper: %s", err.Error())
		return err
	}

	err = newGlobalOpenstackHelper()
	if err != nil {
		log.Errorf("runtime: failed to init openstack helper: %s", err.Error())
		return err
	}

	err = newGlobalKeycloakHelper()
	if err != nil {
		log.Errorf("runtime: failed to init keycloak helper: %s", err.Error())
		return err
	}

	err = newGlobalAwsHelper()
	if err != nil {
		log.Warnf("runtime: failed to init aws helper: %s", err.Error())
	}

	return nil
}

func newAuthIdentities() error {
	err := newKeycloakOidcAuth()
	if err != nil {
		log.Errorf("runtime: failed to init oidc auth in keycloak: %s", err.Error())
		return err
	}

	err = newDefaultOidcSecret()
	if err != nil {
		log.Errorf("runtime: failed to init oidc secret in keycloak: %s", err.Error())
		return err
	}

	err = newDefaultNodeToken()
	if err != nil {
		log.Errorf("runtime: failed to init node token: %s", err.Error())
		return err
	}

	err = newKeycloakSamlMapper()
	if err != nil {
		log.Errorf("runtime: failed to init saml mapper in keycloak: %s", err.Error())
		return err
	}

	v1.NodeMetadata, err = genNodeMetadata()
	if err != nil {
		log.Errorf("runtime: failed to generate node metadata: %s", err.Error())
		return err
	}

	return nil
}

func newSearchIndexers() error {
	err := newTuningSearchIndex()
	if err != nil {
		log.Warnf("runtime: failed to init tuning search index: %s", err.Error())
	}

	err = newSupportFileSearchIndex()
	if err != nil {
		log.Warnf("runtime: failed to init support file search index: %s", err.Error())
	}

	err = newLicenseSearchIndex()
	if err != nil {
		log.Warnf("runtime: failed to init license search index: %s", err.Error())
	}

	err = newNodeSearchIndex()
	if err != nil {
		log.Warnf("runtime: failed to init node search index: %s", err.Error())
	}

	return nil
}

func newReqPendingTTLs() error {
	err := newTuningRecordTTL()
	if err != nil {
		log.Warnf("runtime: failed to init tuning record ttl: %s", err.Error())
	}

	err = newSupportFileRecordTTL()
	if err != nil {
		log.Warnf("runtime: failed to init support file record ttl: %s", err.Error())
	}

	err = newSettingRecordTTL()
	if err != nil {
		log.Warnf("runtime: failed to init setting record ttl: %s", err.Error())
	}

	err = newTriggerRecordTTL()
	if err != nil {
		log.Warnf("runtime: failed to init trigger record ttl: %s", err.Error())
	}

	return nil
}

func newGlobalLogHelper() error {
	opts := conf.Opts.Spec.Observability.Log
	return bslog.NewGlobalHelper(
		bslog.File(opts.File),
		bslog.Level(opts.Level),
		bslog.Backups(opts.Rotation.Backups),
		bslog.Size(opts.Rotation.Size),
		bslog.TTL(opts.Rotation.TTL),
		bslog.Compress(opts.Rotation.Compress),
	)
}

func newGlobalMongoHelper() error {
	opts := conf.Opts.Spec.Store.MongoDB
	return mongo.NewGlobalHelper(
		mongo.Uri(opts.Uri),
		mongo.AuthEnable(opts.Auth.Enable),
		mongo.AuthSource(opts.Auth.Source),
		mongo.AuthUsername(opts.Auth.Username),
		mongo.AuthPassword(opts.Auth.Password),
		mongo.ReplicaSet(opts.ReplicaSet),
	)
}

func newGlobalInfluxHelper() error {
	opts := conf.Opts.Spec.Store.InfluxDB
	return influx.NewGlobalHelper(
		influx.Url(opts.Url),
	)
}

func newGlobalOpenstackHelper() error {
	opts := conf.Opts.Spec.Openstack
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

func newGlobalAwsHelper() error {
	opts := conf.Opts.Spec.Aws
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

func newGlobalKeycloakHelper() error {
	opts := conf.Opts.Spec.Identity.Keycloak
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
		log.Errorf("runtime: failed to login admin: %s", err.Error())
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
		log.Errorf("runtime: failed to login keycloak admin: %s", err.Error())
		return err
	}

	clients, err := h.GetClients(
		v1.DefaultKeycloakRealm,
		gocloak.GetClientsParams{ClientID: gocloak.StringP(v1.DefaultOidcClientId)},
	)
	if err != nil {
		log.Errorf("runtime: failed to get clients: %s", err.Error())
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
		log.Errorf("runtime: failed to get client secret: %s", err.Error())
		return err
	}

	v1.DefaultOidcClientSecret = *secret.Value
	return nil
}

func newDefaultNodeToken() error {
	v1.DefaultNodeToken = v1.GenNodeToken(v1.Hostname)
	if v1.DefaultNodeToken == "" {
		return fmt.Errorf("failed to generate node token")
	}

	return nil
}

func newKeycloakSamlMapper() error {
	url := saml.GenServiceProviderMetadataUrl(conf.Opts.Spec.Identity.Saml)
	client, err := saml.GetSamlClient(url.String())
	if err != nil {
		log.Errorf("runtime: failed to get saml client: %s", err.Error())
		return err
	}

	return saml.CreateSamlMapper(
		*client.ID,
		genSamlMapper(),
	)
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

func newSupportFileSearchIndex() error {
	return support.InitFileSearchIndex()
}

func newLicenseSearchIndex() error {
	return v1.InitLicenseSearchIndex()
}

func newNodeSearchIndex() error {
	return v1.InitNodeSearchIndex()
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

func newSupportFileRecordTTL() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.CreateExpirationIndex(
		support.FileDB,
		support.FileReqCollection,
		bson.D{{Key: "createdAt", Value: 1}},
		support.ReqTTL,
	)
}

func newSettingRecordTTL() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.CreateExpirationIndex(
		setting.DB,
		setting.ReqCollection,
		bson.D{{Key: "createdAt", Value: 1}},
		setting.ReqTTL,
	)
}

func newTriggerRecordTTL() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.CreateExpirationIndex(
		trigger.DB,
		trigger.ReqCollection,
		bson.D{{Key: "createdAt", Value: 1}},
		trigger.ReqTTL,
	)
}
