package integrations

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations",
			Func:    listApplications,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/applications",
			Func:    listApplications,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/storages",
			Func:    listStorages,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/storages/models",
			Func:    listModels,
		},
	}
)

func listApplications(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch integrated applications successfully",
		cubecos.ListBuiltInApplications(),
	)
}

func listStorages(c *gin.Context) {
	h, err := initHelper(c, "listStorages")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	storages, err := h.listStorages()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch integrated storages successfully",
		storages,
	)
}

func listModels(c *gin.Context) {
	h, err := initHelper(c, "listModels")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	models, err := h.listModels()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch integrated models successfully",
		models,
	)
}
