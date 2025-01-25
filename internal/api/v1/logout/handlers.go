package logout

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version:              api.V1,
			Method:               http.MethodPost,
			Path:                 "/logout",
			Func:                 logout,
			IsNotUnderDataCenter: true,
		},
	}
)

func logout(c *gin.Context) {
	session, err := saml.SpAuth.Session.GetSession(c.Request)
	if err != nil {
		log.Errorf("failed to get session for logout: %s", err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = cleanSession(c, session)
	if err != nil {
		api.SetInternalServerError(c, err)
		return
	}

	api.SetRedirect(c, genRedirectUrl())
}
