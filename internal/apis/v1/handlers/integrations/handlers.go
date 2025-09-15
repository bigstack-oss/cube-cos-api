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
			Path:    "/integrations/storages/verify",
			Func:    verifyStorage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/integrations/storages",
			Func:    createStorage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/integrations/storages/:storageName",
			Func:    updateStorage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/integrations/storages/:storageName/asDefault",
			Func:    setStorageAsDefault,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/integrations/storages/:storageName",
			Func:    deleteStorage,
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
			Method:  http.MethodPost,
			Path:    "/integrations/storages/models",
			Func:    createStorageModel,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPut,
			Path:    "/integrations/storages/models",
			Func:    updateAllStorageModels,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPut,
			Path:    "/integrations/storages/models/:vendor/:product",
			Func:    updateStorageModel,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/integrations/storages/models/:vendor/:product",
			Func:    deleteStorageModel,
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

func verifyStorage(c *gin.Context) {
	h, err := initHelper(c, "verifyStorage")
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
		bodies.SetConflict(c, fmt.Errorf("storage %s already exists", h.storageReqOpts.Name))
		return
	}

	verification, err := h.verifyStorage()
	if err != nil {
		return
	}

	bodies.SetOk(
		c,
		"verify integrated storage successfully",
		verification,
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
		"received create integrated storage request successfully, processing",
	)
}

func updateStorage(c *gin.Context) {
	h, err := initHelper(c, "updateStorage")
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

	h.updateStorageToControllers()
	bodies.SetAccepted(
		c,
		"received update integrated storage request successfully, processing",
	)
}

func setStorageAsDefault(c *gin.Context) {
	h, err := initHelper(c, "setStorageAsDefault")
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

	h.updateStorageToControllers()
	bodies.SetAccepted(
		c,
		"received set integrated storage as default request successfully, processing",
	)
}

func deleteStorage(c *gin.Context) {
	h, err := initHelper(c, "deleteStorage")
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

	h.updateStorageToControllers()
	bodies.SetAccepted(
		c,
		"received delete integrated storage request successfully, processing",
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

func createStorageModel(c *gin.Context) {
	h, err := initHelper(c, "createStorageModel")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageModelExist(h.modelReqOpts.Vendor, h.modelReqOpts.Product)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if found {
		bodies.SetConflict(c, fmt.Errorf("storage model '%s %s' already exists", h.modelReqOpts.Vendor, h.modelReqOpts.Product))
		return
	}

	h.updateStorageModelToControllers()
	bodies.SetAccepted(
		c,
		"received create integrated storage model request successfully, processing",
	)
}

func updateAllStorageModels(c *gin.Context) {
	h, err := initHelper(c, "updateAllStorageModels")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.updateAllStorageModelsToControllers()
	bodies.SetAccepted(
		c,
		"received update all integrated storage models request successfully, processing",
	)
}

func updateStorageModel(c *gin.Context) {
	h, err := initHelper(c, "updateStorageModel")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageModelExist(h.modelReqOpts.Vendor, h.modelReqOpts.Product)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage model '%s %s' not found", h.modelReqOpts.Vendor, h.modelReqOpts.Product))
		return
	}

	h.updateStorageModelToControllers()
	bodies.SetAccepted(
		c,
		"received update integrated storage model request successfully, processing",
	)
}

func deleteStorageModel(c *gin.Context) {
	h, err := initHelper(c, "deleteStorageModel")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageModelExist(h.modelReqOpts.Vendor, h.modelReqOpts.Product)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage model '%s %s' not found", h.modelReqOpts.Vendor, h.modelReqOpts.Product))
		return
	}

	h.updateStorageModelToControllers()
	bodies.SetAccepted(
		c,
		"received delete integrated storage model request successfully, processing",
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
