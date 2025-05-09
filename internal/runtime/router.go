package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/healths"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/integrations"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/logout"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/me"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/metrics"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/opensearch"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/services"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/supportfiles"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/tokens"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/triggers"
	apitunings "github.com/bigstack-oss/cube-cos-api/internal/api/v1/tunings"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/oidc"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
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
		server.Name(v1.ServiceDiscoveryIdentity),
		server.Metadata(v1.NodeMetadata),
		server.WithLogger(log.DefaultLogger),
		server.Address(v1.ListenAddr),
		server.Advertise(v1.AdvertiseAddr),
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
		v1.DataCenters,
		datacenters.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Services,
		services.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Me,
		me.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Integrations,
		integrations.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Healths,
		healths.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleCompute,
		v1.RoleStorage,
		v1.RoleModerator,
		v1.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		event.Module,
		events.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Nodes,
		nodes.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleCompute,
		v1.RoleStorage,
		v1.RoleModerator,
		v1.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		v1.Tunings,
		apitunings.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleCompute,
		v1.RoleStorage,
		v1.RoleModerator,
		v1.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		metric.Module,
		metrics.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleCompute,
		v1.RoleStorage,
		v1.RoleModerator,
		v1.RoleEdgeCore,
	)

	api.RegisterHandlersToRoles(
		v1.Tokens,
		tokens.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Logout,
		logout.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		license.Module,
		licenses.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Triggers,
		triggers.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		support.Files,
		supportfiles.Handlers,
		v1.RoleControlConverged,
		v1.RoleControl,
		v1.RoleCompute,
		v1.RoleStorage,
		v1.RoleEdgeCore,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Grafana,
		grafana.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.OpenSearch,
		opensearch.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Settings,
		settings.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
		v1.RoleModerator,
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
		api.SetStatusOk(c, "api is alive", nil)
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
	return token == v1.GenNodeToken(node)
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
	roleHandlers := api.GetRoleHandlers(v1.CurrentRole)
	if len(roleHandlers) == 0 {
		return fmt.Errorf("no handlers found for role(%s)", v1.CurrentRole)
	}

	for _, handlers := range roleHandlers {
		setRoleHandlersToRouter(router, handlers)
	}

	return nil
}
