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
			Method:               http.MethodGet,
			Path:                 "/logout",
			Func:                 logout,
			IsNotUnderDataCenter: true,
		},
	}
)

// @BasePath /api/v1
// @Summary	Logout from the system and redirect to login page
// @Schemes
// @Description
// @Tags		logout      specifications
// @Success	302	{array}     string	""
// @Failure	500	{string}	string	""
// @Router		/logout     [post]
func logout(c *gin.Context) {
	session, err := saml.SpAuth.Session.GetSession(c.Request)
	if err != nil {
		log.Errorf("failed to get session for logout: %s", err.Error())
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	err = cleanSession(c, session)
	if err != nil {
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	api.SetRedirectResp(c, genRedirectUrl())
}
