package logout

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []apis.Handler{
		{
			Version:              apis.V1,
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
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.cleanSession(session)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetRedirect(
		c,
		genRedirectUrl(),
	)
}
