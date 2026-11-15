package runtime

import (
	"errors"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/authz"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

// authorize checks that the authenticated caller may call the route, as
// declared by the route's Access.
func authorize(h apis.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch h.Access {
		case apis.AccessHandler:
			return
		case apis.AccessNode:
			if !c.GetBool("isNode") {
				refuse(c, "only cluster nodes may call this route")
			}
		case apis.AccessAdmin:
			authorizeAdmin(c)
		default:
			if c.Request.Method != http.MethodGet {
				log.Errorf("authz(%s): %s %s declares no access", queries.GetReqId(c), c.Request.Method, c.FullPath())
				refuse(c, "this route is not open to any caller")
			}
		}
	}
}

func authorizeAdmin(c *gin.Context) {
	caller, err := authz.Resolve(c)
	if errors.Is(err, authz.ErrUnavailable) {
		bodies.SetServiceUnavailable(c, err)
		c.Abort()
		return
	}

	if err != nil || !caller.IsAdmin {
		refuse(c, "only cluster admins may call this route")
	}
}

func refuse(c *gin.Context, reason string) {
	log.Warnf(
		"authz(%s): refused %s %s for %s caller %s(%s)",
		queries.GetReqId(c), c.Request.Method, c.FullPath(), c.GetString("authType"), callerName(c), reason,
	)

	bodies.SetForbidden(c, errors.New(reason))
	c.Abort()
}

// callerName names the caller in the log. SAML requests carry the user in
// the session rather than in authUser.
func callerName(c *gin.Context) string {
	if c.GetString("authType") == auths.Saml {
		return samlsp.AttributeFromContext(c.Request.Context(), "username")
	}

	return c.GetString("authUser")
}
