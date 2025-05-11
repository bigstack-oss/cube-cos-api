package tokens

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/tokens",
			Func:    createToken,
		},
	}
)

func createToken(c *gin.Context) {
	user, err := parseUserBody(c)
	if err != nil {
		log.Errorf("tokens(%s): failed to parse user: %s", queries.GetReqId(c), err.Error())
		bodies.SetBadRequest(c, err)
		return
	}

	auth, err := cubecos.CreateToken(user)
	if err != nil {
		log.Errorf("tokens(%s): failed to generate token: %s", queries.GetReqId(c), err.Error())
		bodies.SetUnauthorized(c, err)
		return
	}

	bodies.SetCreated(
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
