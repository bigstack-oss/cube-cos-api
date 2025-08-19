package fixpacks

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/fixpacks",
			Func:    listFixpacks,
		},
	}
)

func listFixpacks(c *gin.Context) {
	h, err := initHelper(c, "listFixpacks")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	fixpacks, err := h.listFixpacks()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of fixpacks",
		fixpacks,
	)
}
