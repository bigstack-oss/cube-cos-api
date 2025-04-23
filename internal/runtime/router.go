package runtime

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/oidc"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/micro/plugins/v5/server/http"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
)

func newHttpServer() (*server.Server, error) {
	router := newRouter()
	err := registerHandlersByRole(router)
	if err != nil {
		log.Errorf("runtime: failed to register handlers: %s", err.Error())
		return nil, err
	}

	srv := http.NewServer(
		server.Name(v1.DataCenterName),
		server.Metadata(v1.NodeMetadata),
		server.WithLogger(log.DefaultLogger),
		server.Address(v1.ListenAddr),
		server.Advertise(v1.AdvertiseAddr),
	)

	err = srv.Handle(srv.NewHandler(router))
	if err != nil {
		log.Errorf("runtime: failed to new handler: %s", err.Error())
		return nil, err
	}

	return &srv, nil
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(initReqInfo)
	router.Any("/live", livenessCheck())
	router.Any("/saml/*any", saml.ServeAcs())
	router.Use(verifyAuthToken())
	router.Use(conditionalSaml())
	return router
}

func livenessCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		api.SetStatusOk(c, "api is alive", nil)
	}
}

func initReqInfo(c *gin.Context) {
	reqId := uuid.New().String()[:8]
	c.Set("reqId", reqId)
	log.Infof("request(%s): %s %s%s", reqId, c.Request.Method, c.Request.URL.Path, parseParams(c))
	c.Next()
}

func verifyAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := parseToken(c)
		if token == v1.DefaultNodeToken {
			c.Set("isTokenValid", true)
			c.Set("authType", "oidc")
			c.Set("authUser", c.ClientIP())
			c.Next()
			return
		}

		claims, err := oidc.VerifyToken(token)
		if err == nil {
			c.Set("isTokenValid", true)
			c.Set("authType", "oidc")
			c.Set("authUser", claims.PreferredUsername)
		}

		c.Next()
	}
}

func parseToken(c *gin.Context) string {
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

func registerHandlersByRole(router *gin.Engine) error {
	groupHandlers := api.GetGroupHandlersByRole(v1.CurrentRole)
	if len(groupHandlers) == 0 {
		return fmt.Errorf("no handlers found for role(%s)", v1.CurrentRole)
	}

	for _, handlers := range groupHandlers {
		setGroupHandlersToRouter(router, handlers)
	}

	return nil
}

func setGroupHandlersToRouter(router *gin.Engine, handlers []api.Handler) {
	for _, h := range handlers {
		if h.Version == "" {
			log.Warnf("runtime: skip invalid API registration: %s %s (no version or controller provided)", h.Method, h.Path)
			continue
		}

		parentPath := getParentPath(h)
		routerGroup := router.Group(parentPath)
		routerGroup.Handle(h.Method, h.Path, h.Func)
		log.Infof("register API: %s %s", h.Method, fmt.Sprintf("%s%s", parentPath, h.Path))
	}
}

func getParentPath(h api.Handler) string {
	if h.IsNotUnderDataCenter {
		return h.Version
	}

	return fmt.Sprintf("%s/datacenters/:DataCenter", h.Version)
}
