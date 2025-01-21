package runtime

import (
	"fmt"
	"os"

	apihttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/log"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/health"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/integrations"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/logout"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/summary"
	apitunings "github.com/bigstack-oss/cube-cos-api/internal/api/v1/tunings"
	apiConf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/controllers/v1/node"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/micro/plugins/v5/server/http"
	"go-micro.dev/v5/config"
	"go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
)

func NewRuntime(conf config.Config) (*server.Server, error) {
	err := conf.Get().Scan(&apiConf.Opts)
	if err != nil {
		logger.Errorf("failed to scan config: %s", err.Error())
		return nil, err
	}

	err = initNodeIdentities()
	if err != nil {
		logger.Errorf("failed to init node identities: %s", err.Error())
		return nil, err
	}

	err = initDependencyHelpers()
	if err != nil {
		logger.Errorf("failed to init node clis: %s", err.Error())
		return nil, err
	}

	initNodePeerSyncer()
	initNodeApiHandler()

	showPromptMessages()
	showLoadedConfBody()

	return newHttpServer()
}

func initNodeIdentities() error {
	var err error
	definition.Hostname, err = os.Hostname()
	if err != nil {
		logger.Errorf("failed to get hostname: %s", err.Error())
		return err
	}

	definition.HostID, err = definition.GenerateNodeHashByMacAddr()
	if err != nil {
		logger.Errorf("failed to generate host id: %s", err.Error())
		return err
	}

	definition.CurrentRole, err = cubecos.GetNodeRole()
	if err != nil {
		logger.Errorf("failed to get node role: %s", err.Error())
		return err
	}

	definition.IsHaEnabled, err = cubecos.IsHaEnabled()
	if err != nil {
		logger.Errorf("failed to get ha enabled: %s", err.Error())
		return err
	}

	definition.MgmtNet, err = cubecos.GetMgmtNet()
	if err != nil {
		logger.Errorf("failed to get management network: %s", err.Error())
		return err
	}

	definition.MgmtIP, err = cubecos.GetManagementIp(definition.MgmtNet)
	if err != nil {
		logger.Errorf("failed to get management ip: %s", err.Error())
		return err
	}

	definition.ControllerVip, err = cubecos.GetControllerVirtualIp(definition.MgmtNet)
	if err != nil {
		logger.Errorf("failed to get controller virtual ip: %s", err.Error())
		return err
	}

	definition.Controller, err = cubecos.GetDataCenterName()
	if err != nil {
		logger.Errorf("failed to get data center name: %s", err.Error())
		return err
	}

	definition.ListenAddr = genLocalAddr()
	definition.AdvertiseAddr = genServiceDiscoveryAddr()
	definition.IsGpuEnabled = cubecos.IsGpuEnabled()
	definition.LogoutRedirectUrl = genLogoutRedirectUrl()

	return nil
}

func initDependencyHelpers() error {
	err := newGlobalLogHelper(apiConf.Opts.Spec.Observability.Log)
	if err != nil {
		logger.Errorf("failed to init logger: %s", err.Error())
		return err
	}

	err = newGlobalHttpHelper()
	if err != nil {
		logger.Errorf("failed to init http helper: %s", err.Error())
		return err
	}

	err = newGlobalSamlHelper()
	if err != nil {
		logger.Errorf("failed to init keycloak auth: %s", err.Error())
		return err
	}

	err = newGlobalMongoHelper(apiConf.Opts.Spec.Store.MongoDB)
	if err != nil {
		logger.Errorf("failed to init mongo helper: %s", err.Error())
		return err
	}

	err = newGlobalInfluxHelper(apiConf.Opts.Spec.Store.InfluxDB)
	if err != nil {
		logger.Errorf("failed to init influx helper: %s", err.Error())
		return err
	}

	err = newGlobalOpenstackHelper(apiConf.Opts.Spec.Openstack)
	if err != nil {
		logger.Errorf("failed to init openstack helper: %s", err.Error())
		return err
	}

	err = newGlobalKeycloakHelper(apiConf.Opts.Spec.Identity.Keycloak)
	if err != nil {
		logger.Errorf("failed to init keycloak helper: %s", err.Error())
		return err
	}

	return nil
}

func initNodePeerSyncer() {
	service.RegisterController(node.Name(), node.NewController())
}

func initNodeApiHandler() {
	api.RegisterHandlersToRoles(
		definition.DataCenters,
		datacenters.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Summary,
		summary.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Integrations,
		integrations.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Health,
		health.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Events,
		events.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Nodes,
		nodes.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Tunings,
		apitunings.Handlers,
		definition.RoleControl,
		definition.RoleCompute,
	)

	api.RegisterHandlersToRoles(
		definition.Logout,
		logout.Handlers,
		definition.RoleControl,
	)
}

func newGlobalLogHelper(opts log.Options) error {
	return log.NewGlobalHelper(
		log.File(opts.File),
		log.Level(opts.Level),
		log.Backups(opts.Rotation.Backups),
		log.Size(opts.Rotation.Size),
		log.TTL(opts.Rotation.TTL),
		log.Compress(opts.Rotation.Compress),
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
	return apihttp.NewGlobalHelper()
}

func newGlobalSamlHelper() error {
	return saml.NewGlobalAuth(apiConf.Opts.Spec.Identity.Saml)
}

func newHttpServer() (*server.Server, error) {
	router := newRouter()
	err := registerHandlersByRole(router)
	if err != nil {
		logger.Errorf("failed to register handlers: %s", err.Error())
		return nil, err
	}

	srv := http.NewServer(
		server.Name(definition.CurrentRole),
		server.Metadata(genMetadata()),
		server.WithLogger(logger.DefaultLogger),
		server.Address(genLocalAddr()),
		server.Advertise(genServiceDiscoveryAddr()),
	)

	err = srv.Handle(srv.NewHandler(router))
	if err != nil {
		logger.Errorf("failed to new handler: %s", err.Error())
		return nil, err
	}

	return &srv, nil
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Any("/saml/*any", gin.WrapH(saml.SpAuth))
	router.Use(gin.Recovery())
	router.Use(initReqInfo)
	router.Use(adapter.Wrap(saml.SpAuth.RequireAccount))
	return router
}

func initReqInfo(c *gin.Context) {
	uuidV4 := uuid.New()
	c.Set("reqId", uuidV4.String())
	logger.Infof("request(%s): %s %s", uuidV4, c.Request.Method, c.Request.URL.Path)
	c.Next()
}

func genLocalAddr() string {
	return fmt.Sprintf(
		"%s:%d",
		apiConf.Opts.Spec.Listen.Local,
		apiConf.Opts.Spec.Listen.Port,
	)
}

func genServiceDiscoveryAddr() string {
	return fmt.Sprintf(
		"%s:%d",
		definition.MgmtIP,
		apiConf.Opts.Spec.Listen.Port,
	)
}

func genLogoutRedirectUrl() string {
	return fmt.Sprintf(
		"https://%s:4443%s",
		definition.ControllerVip,
		apiConf.Opts.Spec.Identity.LogoutRedirect,
	)
}

func genMetadata() map[string]string {
	return map[string]string{
		"hostname":     definition.Hostname,
		"nodeID":       definition.HostID,
		"isGpuEnabled": fmt.Sprintf("%t", definition.IsGpuEnabled),
	}
}

func registerHandlersByRole(router *gin.Engine) error {
	groupHandlers := api.GetGroupHandlersByRole(definition.CurrentRole)
	if len(groupHandlers) == 0 {
		return fmt.Errorf("no handlers found for role(%s)", definition.CurrentRole)
	}

	for _, handlers := range groupHandlers {
		setGroupHandlersToRouter(router, handlers)
	}

	return nil
}

func setGroupHandlersToRouter(router *gin.Engine, handlers []api.Handler) {
	for _, h := range handlers {
		if h.Version == "" {
			logger.Warnf("skip invalid API registration: %s %s (no version or controller provided)", h.Method, h.Path)
			continue
		}

		parentPath := getParentPath(h)
		routerGroup := router.Group(parentPath)
		routerGroup.Handle(h.Method, h.Path, h.Func)
		logger.Infof("register API: %s %s", h.Method, fmt.Sprintf("%s%s", parentPath, h.Path))
	}
}

func getParentPath(h api.Handler) string {
	if h.IsNotUnderDataCenter {
		return h.Version
	}

	return fmt.Sprintf("%s/datacenters/:DataCenter", h.Version)
}
