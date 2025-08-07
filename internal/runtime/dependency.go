package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/aws"
	bshttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/kubernetes"
	bslog "github.com/bigstack-oss/bigstack-dependency-go/pkg/log"
	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/saml"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/gophercloud/gophercloud/v2"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	err = newDbHousekeeping()
	if err != nil {
		return err
	}

	err = newDryRunFacility()
	if err != nil {
		return err
	}

	return nil
}

func newGlobalHelpers() error {
	err := newGlobalLogHelper()
	if err != nil {
		log.Errorf("runtime: failed to init logger(%v)", err)
		return err
	}

	err = newGlobalHttpHelper()
	if err != nil {
		log.Errorf("runtime: failed to init http helper(%v)", err)
		return err
	}

	err = newGlobalSamlHelper()
	if err != nil {
		log.Errorf("runtime: failed to init keycloak auth(%v)", err)
		return err
	}

	err = newGlobalMongoHelper()
	if err != nil {
		log.Errorf("runtime: failed to init mongo helper(%v)", err)
		return err
	}

	err = newGlobalInfluxHelper()
	if err != nil {
		log.Errorf("runtime: failed to init influx helper(%v)", err)
		return err
	}

	err = newGlobalOpenstackHelper()
	if err != nil {
		log.Errorf("runtime: failed to init openstack helper(%v)", err)
		return err
	}

	err = newGlobalKubernetesHelper()
	if err != nil {
		log.Errorf("runtime: failed to init kubernetes helper(%v)", err)
	}

	err = newGlobalKeycloakHelper()
	if err != nil {
		log.Errorf("runtime: failed to init keycloak helper(%v)", err)
		return err
	}

	err = newKeycloakOidcAuth()
	if err != nil {
		log.Errorf("runtime: failed to init oidc auth in keycloak(%v)", err)
		return err
	}

	err = newDefaultOidcSecret()
	if err != nil {
		log.Errorf("runtime: failed to init oidc secret in keycloak(%v)", err)
		return err
	}

	err = newGlobalAwsHelper()
	if err != nil {
		log.Warnf("runtime: failed to init aws helper(%v)", err)
	}

	return nil
}

func newAuthIdentities() error {
	err := newDefaultNodeToken()
	if err != nil {
		log.Errorf("runtime: failed to init node token(%v)", err)
		return err
	}

	err = newKeycloakSamlMapper()
	if err != nil {
		log.Errorf("runtime: failed to init saml mapper in keycloak(%v)", err)
		return err
	}

	err = newBucketSecret()
	if err != nil {
		log.Errorf("runtime: failed to init bucket secret(%v)", err)
		return err
	}

	err = newServiceDiscoveryIdentity()
	if err != nil {
		log.Errorf("runtime: failed to parse service discovery identify(%v)", err)
		return err
	}

	err = newNodeMetadata()
	if err != nil {
		log.Errorf("runtime: failed to generate node metadata(%v)", err)
		return err
	}

	return nil
}

func newDbHousekeeping() error {
	helper := bsmongo.GetGlobalHelper()
	cli, err := helper.NewCollCli(
		notifications.Db,
		notifications.ToastCollection,
	)
	if err != nil {
		log.Errorf("runtime: failed to create mongo client for ttl setup(%v)", err)
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	oneMonth := int32(3600 * 24 * 30)
	ttlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "time", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(oneMonth),
	}
	_, err = cli.Indexes().CreateOne(ctx, ttlIndex)
	if err == nil {
		return nil
	}

	if ceph.IsIndexExistsError(err) {
		log.Warnf("runtime: ttl index already exists for %s(%v)", notifications.ToastCollection, err)
		return nil
	}

	return nil
}

func newDryRunFacility() error {
	err := newTriggerDryRunFacility()
	if err != nil {
		log.Errorf("runtime: failed to create dry run facility(%v)", err)
		return err
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
	opts := parseMongoOpts()
	return bsmongo.NewGlobalHelper(
		bsmongo.Uri(opts.Uri),
		bsmongo.AuthEnable(opts.Auth.Enable),
		bsmongo.AuthSource(opts.Auth.Source),
		bsmongo.AuthUsername(opts.Auth.Username),
		bsmongo.AuthPassword(opts.Auth.Password),
		bsmongo.ReplicaSet(opts.ReplicaSet),
	)
}

func newGlobalInfluxHelper() error {
	opts := parseInfluxOpts()
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

func newGlobalKubernetesHelper() error {
	return kubernetes.NewGlobalHelper(
		kubernetes.AuthType(kubernetes.OutOfClusterAuth),
		kubernetes.AuthFile(conf.Opts.Spec.ResourceControl.K3s.Auth),
	)
}

func newGlobalAwsHelper() error {
	opts := conf.Opts.Spec.Aws
	if opts.SecretKey == "" {
		conf.Opts.Spec.Aws.SecretKey = auths.DefaultOidcClientSecret
		opts.SecretKey = auths.DefaultOidcClientSecret
	}

	return aws.NewGlobalHelper(
		aws.Region(opts.Region),
		aws.EnableStaticCreds(opts.EnableStaticCreds),
		aws.AccessKey(opts.AccessKey),
		aws.SecretKey(opts.SecretKey),
		aws.EnableCustomURL(opts.EnableCustomURL),
		aws.S3Url(ceph.GetRadosGatewayUrl()),
		aws.InsecureSkipVerify(opts.InsecureSkipVerify),
	)
}

func newGlobalKeycloakHelper() error {
	opts := conf.Opts.Spec.Identity.Keycloak
	if opts.Ip == "" {
		opts.Ip = base.DataCenterVip
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
	return bshttp.NewGlobalHelper(
		bshttp.Timeout(180 * time.Second),
	)
}

func newGlobalSamlHelper() error {
	return saml.NewGlobalAuth(conf.Opts.Spec.Identity.Saml)
}

func newKeycloakOidcAuth() error {
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
	if err != nil {
		log.Errorf("runtime: failed to login admin(%v)", err)
		return err
	}

	_, err = h.CreateClient(
		auths.DefaultKeycloakRealm,
		auths.DefaultOidcClientOpts,
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
		log.Errorf("runtime: failed to login keycloak admin(%v)", err)
		return err
	}

	clients, err := h.GetClients(
		auths.DefaultKeycloakRealm,
		gocloak.GetClientsParams{ClientID: gocloak.StringP(auths.DefaultOidcClientId)},
	)
	if err != nil {
		log.Errorf("runtime: failed to get clients(%v)", err)
		return err
	}
	if len(clients) == 0 {
		return fmt.Errorf("oidc client not found")
	}

	secret, err := h.GetClientSecret(
		auths.DefaultKeycloakRealm,
		*clients[0].ID,
	)
	if err != nil {
		log.Errorf("runtime: failed to get client secret(%v)", err)
		return err
	}

	auths.DefaultOidcClientSecret = *secret.Value
	return nil
}

func newDefaultNodeToken() error {
	auths.DefaultNodeToken = nodes.GenToken(base.Hostname)
	if auths.DefaultNodeToken == "" {
		return fmt.Errorf("failed to generate node token")
	}

	return nil
}

func newKeycloakSamlMapper() error {
	url := saml.GenServiceProviderMetadataUrl(conf.Opts.Spec.Identity.Saml)
	client, err := saml.GetSamlClient(url.String())
	if err != nil {
		log.Errorf("runtime: failed to get saml client(%v)", err)
		return err
	}

	return saml.CreateSamlMapper(
		*client.ID,
		genSamlMapper(),
	)
}

func newBucketSecret() error {
	h := openstack.GetGlobalHelper()
	accessKey := conf.Opts.Spec.Aws.AccessKey
	secretKey := conf.Opts.Spec.Aws.SecretKey
	if secretKey == "" {
		secretKey = auths.DefaultOidcClientSecret
	}

	userId, err := h.GetUserIdByName(accessKey)
	if err != nil {
		return err
	}

	projectId, err := h.GetProjectIdByName(auths.DefaultAdminProject)
	if err != nil {
		return err
	}

	_, err = h.CreateEc2Credential(userId, projectId, accessKey, secretKey)
	if err == nil {
		return nil
	}
	if gophercloud.ResponseCodeIs(err, http.StatusConflict) {
		return nil
	}

	return err
}

func genSamlMapper() gocloak.ProtocolMapperRepresentation {
	return gocloak.ProtocolMapperRepresentation{
		Name:           gocloak.StringP("username"),
		Protocol:       gocloak.StringP(auths.Saml),
		ProtocolMapper: gocloak.StringP("saml-user-property-mapper"),
		Config: &map[string]string{
			"user.attribute":       "username",
			"attribute.name":       "username",
			"attribute.nameformat": "Basic",
		},
	}
}

func newTriggerDryRunFacility() error {
	h := kubernetes.GetGlobalHelper()
	h.SetNamespaceClient()
	err := h.CreateNamespace(metav1.ObjectMeta{
		Name: triggers.DryRunNamespace,
	})
	if err == nil {
		return nil
	}

	if errors.IsAlreadyExists(err) {
		return nil
	}

	return fmt.Errorf(
		"failed to create trigger script dry run namespace(%v)",
		err,
	)
}
