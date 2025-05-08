package logout

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
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
	h := initHelper(c)
	session, err := h.getSession()
	if err != nil {
		api.SetInternalServerError(c, err)
		return
	}

	err = h.cleanSession(session)
	if err != nil {
		api.SetInternalServerError(c, err)
		return
	}

	api.SetRedirect(c, genRedirectUrl())
}
