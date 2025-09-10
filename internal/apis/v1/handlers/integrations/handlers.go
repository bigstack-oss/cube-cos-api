package integrations

import (
	"fmt"
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
			Method:  http.MethodPost,
			Path:    "/integrations/storages",
			Func:    createStorage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/storages/:storageName",
			Func:    getStorage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/storages/vendors",
			Func:    listVendors,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/storages/models",
			Func:    listModels,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/integrations/storages/tasks",
			Func:    updateStorageTask,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/integrations/storages/models/tasks",
			Func:    updateModelTask,
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

func getStorage(c *gin.Context) {
	h, err := initHelper(c, "getStorage")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage %s not found", h.storageReqOpts.Name))
		return
	}

	storage, err := cubecos.GetStorage(h.storageReqOpts.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch integrated storage details successfully",
		storage,
	)
}

func createStorage(c *gin.Context) {
	h, err := initHelper(c, "createStorage")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if found {
		bodies.SetNotFound(c, fmt.Errorf("storage %s already exists", h.storageReqOpts.Name))
		return
	}

	h.updateStorageToControllers()
	bodies.SetAccepted(
		c,
		"create integrated storage successfully",
	)
}

func listVendors(c *gin.Context) {
	h, err := initHelper(c, "listVendors")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	vendors, err := h.listVendors()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch integrated storage vendors successfully",
		vendors,
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

func updateStorageTask(c *gin.Context) {
	h, err := initHelper(c, "updateStorageTask")
	if err != nil {
		log.Errorf("storages(%s): failed to init request helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkTaskUpdateReq()
	if err != nil {
		log.Errorf("storages(%s): failed to check storage task(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateStorageTask()
	if err != nil {
		log.Errorf("storages(%s): failed to update storage status(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"storage status updated",
		nil,
	)
}

func updateModelTask(c *gin.Context) {
	h, err := initHelper(c, "updateModelTask")
	if err != nil {
		log.Errorf("storages(%s): failed to init request helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkTaskUpdateReq()
	if err != nil {
		log.Errorf("storages(%s): failed to check model task(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateModelTask()
	if err != nil {
		log.Errorf("storages(%s): failed to update model status(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"model status updated",
		nil,
	)
}
