package images

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/images/materials",
			Func:    listImageMaterials,
		},
	}
)

func listImageMaterials(c *gin.Context) {
	h, err := initHelper(c, "listMaterials")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	materials, err := h.listMaterials()
	if err != nil {
		log.Errorf("images(%s): failed to list materials(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch image materials successfully",
		materials,
	)
}
