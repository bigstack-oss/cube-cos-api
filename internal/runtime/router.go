package runtime

import (
	"fmt"
	"strings"
	"time"

	api "github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	datacenterapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/events"
	grafanapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/healths"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/integrations"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/logout"
	meapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/me"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/metrics"
	nodeapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/nodes"
	opensearchapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/opensearch"
	servicesapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/services"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/supportfiles"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/tokens"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/triggers"
	tuningapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/tunings"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/oidc"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/saml"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/health"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/me"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/opensearch"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/services"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/micro/plugins/v5/server/http"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
)

func newHttpServer() (*server.Server, error) {
	router := newGinRouter()
	prepareApiHandleraByRole()
	err := registerHandlersByCurrentRole(router)
	if err != nil {
		log.Errorf("runtime: failed to register handlers: %s", err.Error())
		return nil, err
	}

	srv := http.NewServer(
		server.Name(base.ServiceDiscoveryIdentity),
		server.Metadata(base.NodeMetadata),
		server.WithLogger(log.DefaultLogger),
		server.Address(base.ListenAddr),
		server.Advertise(base.AdvertiseAddr),
	)

	return &srv,
		srv.Handle(srv.NewHandler(router))
}

func newGinRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(timeTracker)
	router.Use(initReqInfo)
	router.Any("/live", checkLive())
	router.Any("/saml/*any", saml.ServeAcs())
	router.Use(verifyAuthToken())
	router.Use(conditionalSaml())
	return router
}

func setRoleHandlersToRouter(router *gin.Engine, handlers []api.Handler) {
	for _, h := range handlers {
		if h.Version == "" {
			log.Warnf("runtime: skip invalid API registration: %s %s(no version provided)", h.Method, h.Path)
			continue
		}

		urlParentPath := getUrlParentPath(h)
		versionGroup := router.Group(urlParentPath)
		versionGroup.Handle(h.Method, h.Path, h.Func)
		log.Infof("register API: %s %s", h.Method, fmt.Sprintf("%s%s", urlParentPath, h.Path))
	}
}

func getUrlParentPath(h api.Handler) string {
	if h.IsNotUnderDataCenter {
		return h.Version
	}

	return fmt.Sprintf("%s/datacenters/:DataCenter", h.Version)
}

func prepareApiHandleraByRole() {
	api.RegisterHandlersToRoles(
		base.DataCenters,
		datacenterapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		services.ModuleName,
		servicesapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		me.Module,
		meapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		integration.Module,
		integrations.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		health.Module,
		healths.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleCompute,
		nodes.RoleStorage,
		nodes.RoleModerator,
		nodes.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		event.Module,
		events.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		nodes.Module,
		nodeapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleCompute,
		nodes.RoleStorage,
		nodes.RoleModerator,
		nodes.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		tunings.Module,
		tuningapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleCompute,
		nodes.RoleStorage,
		nodes.RoleModerator,
		nodes.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		metric.Module,
		metrics.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleCompute,
		nodes.RoleStorage,
		nodes.RoleModerator,
		nodes.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		auths.Tokens,
		tokens.Handlers,
		nodes.RoleControl,
	)

	api.RegisterHandlersToRoles(
		auths.Logout,
		logout.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		license.Module,
		licenses.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		trigger.Module,
		triggers.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		support.Files,
		supportfiles.Handlers,
		nodes.RoleControlConverged,
		nodes.RoleControl,
		nodes.RoleCompute,
		nodes.RoleStorage,
		nodes.RoleEdgeCore,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		grafana.Module,
		grafanapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		opensearch.Module,
		opensearchapi.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		setting.Module,
		settings.Handlers,
		nodes.RoleControl,
		nodes.RoleControlConverged,
		nodes.RoleModerator,
	)
}

func timeTracker(c *gin.Context) {
	start := time.Now()
	c.Next()
	elapsed := time.Since(start)
	reqId, found := c.Get("reqId")
	if !found {
		reqId = "unknown"
	}

	log.Infof(
		"req(%s): %s (%s)",
		reqId,
		genRequestMsg(c),
		elapsed,
	)
}

func initReqInfo(c *gin.Context) {
	reqId := uuid.New().String()[:8]
	c.Set("reqId", reqId)
	c.Next()
}

func checkLive() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodies.SetOk(c, "api is alive", nil)
	}
}

func verifyAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAuthFreeReq(c) {
			c.Set("isTokenValid", true)
			c.Set("authType", "none")
			c.Set("authUser", c.ClientIP())
			c.Next()
			return
		}

		internalToken := parseInternalToken(c)
		if isValidInternalToken(c, internalToken) {
			c.Set("isTokenValid", true)
			c.Set("authType", "oidc")
			c.Set("authUser", c.ClientIP())
			c.Next()
			return
		}

		oidcToken := parseOidcToken(c)
		claims, err := oidc.VerifyToken(oidcToken)
		if err == nil {
			c.Set("isTokenValid", true)
			c.Set("authType", "oidc")
			c.Set("authUser", claims.PreferredUsername)
		}

		c.Next()
	}
}

func isAuthFreeReq(c *gin.Context) bool {
	return c.Request.Method == "GET" && c.Request.URL.Path == "/api/v1/datacenters"
}

func parseInternalToken(c *gin.Context) string {
	node := c.GetHeader("Node")
	if node == "" {
		return ""
	}

	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	const bearer = "Bearer "
	if !strings.HasPrefix(auth, bearer) {
		return ""
	}

	return strings.TrimPrefix(auth, bearer)
}

func parseOidcToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	const bearer = "Bearer "
	if !strings.HasPrefix(auth, bearer) {
		return ""
	}

	return strings.TrimPrefix(auth, bearer)
}

func isValidInternalToken(c *gin.Context, token string) bool {
	node := c.GetHeader("Node")
	return token == nodes.GenToken(node)
}

func conditionalSaml() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, found := c.Get("isTokenValid")
		if found {
			c.Next()
			return
		}

		if isTokenRequest(c) {
			c.Next()
			return
		}

		c.Set("authType", "saml")
		saml.AuthRequest(c)
	}
}

func isTokenRequest(c *gin.Context) bool {
	return strings.Contains(c.Request.URL.Path, "/token")
}

func registerHandlersByCurrentRole(router *gin.Engine) error {
	roleHandlers := api.GetRoleHandlers(base.CurrentRole)
	if len(roleHandlers) == 0 {
		return fmt.Errorf("no handlers found for role(%s)", base.CurrentRole)
	}

	for _, handlers := range roleHandlers {
		setRoleHandlersToRouter(router, handlers)
	}

	return nil
}
