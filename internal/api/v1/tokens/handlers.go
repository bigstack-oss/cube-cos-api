package tokens

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodPost,
			Path:    "/tokens",
			Func:    createToken,
		},
	}
)

func createToken(c *gin.Context) {
	user, err := parseUserBody(c)
	if err != nil {
		log.Infof("failed to parse user info: %s", err.Error())
		api.SetBadRequest(c, err)
		return
	}

	auth, err := cubecos.CreateToken(user)
	if err != nil {
		log.Infof("failed to generate token: %s", err.Error())
		api.SetUnauthorized(c, err)
		return
	}

	api.SetStatusCreated(
		c,
		"create token successfully",
		token{
			Access:  auth.AccessToken,
			Refresh: auth.RefreshToken,
			Expires: expires{
				Access:  auth.ExpiresIn,
				Refresh: auth.RefreshExpiresIn,
			},
		},
	)
}
