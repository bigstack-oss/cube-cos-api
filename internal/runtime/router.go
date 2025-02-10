package runtime

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/oidc"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/micro/plugins/v5/server/http"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
)

func newHttpServer() (*server.Server, error) {
	router := newRouter()
	err := registerHandlersByRole(router)
	if err != nil {
		log.Errorf("failed to register handlers: %s", err.Error())
		return nil, err
	}

	srv := http.NewServer(
		server.Name(definition.DataCenterName),
		server.Metadata(genMetadata()),
		server.WithLogger(log.DefaultLogger),
		server.Address(genLocalAddr()),
		server.Advertise(genServiceDiscoveryAddr()),
	)

	err = srv.Handle(srv.NewHandler(router))
	if err != nil {
		log.Errorf("failed to new handler: %s", err.Error())
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
	router.Use(verifyDevSecret())
	router.Use(verifyAuthToken())
	router.Use(conditionalSaml())
	return router
}

func initReqInfo(c *gin.Context) {
	uuidV4 := uuid.New()
	c.Set("reqId", uuidV4.String())
	log.Infof("request(%s): %s %s", uuidV4, c.Request.Method, c.Request.URL.Path)
	c.Next()
}

// TODO M1: Due to we still don't have a mechanism for API communicate
// so use a dev secret for it when developing, but should be removed before M1 release
func verifyDevSecret() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("secret")
		if secret == "Dev@Cube" {
			c.Set("isSecretValid", true)
		}

		c.Next()
	}
}

func verifyAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, found := c.Get("isSecretValid")
		if found {
			c.Next()
			return
		}

		token := parseToken(c)
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
	doSamlAuth := adapter.Wrap(saml.SpAuth.RequireAccount)
	return func(c *gin.Context) {
		_, found := c.Get("isSecretValid")
		if found {
			c.Next()
			return
		}

		_, found = c.Get("isTokenValid")
		if found {
			c.Next()
			return
		}

		// M1 TODO: can converge with the saml auth
		if strings.Contains(c.Request.URL.Path, "/token") {
			c.Next()
			return
		}

		c.Set("authType", "saml")
		doSamlAuth(c)
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
			log.Warnf("skip invalid API registration: %s %s (no version or controller provided)", h.Method, h.Path)
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
