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
			Path:    "/integrations/storages/:storageName/verify",
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
			Method:  http.MethodPatch,
			Path:    "/integrations/storages/models/:driverName",
			Func:    updateStorageModel,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPut,
			Path:    "/integrations/storages/models",
			Func:    updateStorageModels,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/integrations/storages/models/:driverName",
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

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage %s not found", h.storageReqOpts.CinderDetails.Name))
		return
	}

	storage, err := cubecos.GetStorage(h.storageReqOpts.CinderDetails.Name)
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

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetConflict(c, fmt.Errorf("storage %s does not exist", h.storageReqOpts.CinderDetails.Name))
		return
	}

	h.requestStorageVerification()
	bodies.SetAccepted(
		c,
		"verify integrated storage successfully, processing",
	)
}

func createStorage(c *gin.Context) {
	h, err := initHelper(c, "createStorage")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if found {
		bodies.SetConflict(c, fmt.Errorf("storage %s already exists", h.storageReqOpts.CinderDetails.Name))
		return
	}

	h.updateStorage()
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

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage %s not found", h.storageReqOpts.CinderDetails.Name))
		return
	}

	h.updateStorage()
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

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage %s not found", h.storageReqOpts.CinderDetails.Name))
		return
	}

	if !h.hasVerifiedRecord(h.storageReqOpts.CinderDetails.Name) {
		bodies.SetBadRequest(c, fmt.Errorf("storage %s is not verified yet", h.storageReqOpts.CinderDetails.Name), nil)
		return
	}

	err = h.checkIfStorageIsDefaulted()
	if err != nil {
		return
	}

	h.updateStorage()
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

	found, err := cubecos.CheckStorageExist(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage %s not found", h.storageReqOpts.CinderDetails.Name))
		return
	}

	h.updateStorage()
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

	err = h.checkStorageTaskUpdateReq()
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

	h.updatePeerStorageModel()
	bodies.SetAccepted(
		c,
		"received create integrated storage model request successfully, processing",
	)
}

func updateStorageModel(c *gin.Context) {
	h, err := initHelper(c, "updateStorageModel")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageModelExist(h.modelReqOpts.Driver)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage model '%s' not found", h.modelReqOpts.Driver))
		return
	}

	h.updatePeerStorageModel()
	bodies.SetAccepted(
		c,
		"received update integrated storage model request successfully, processing",
	)
}

func updateStorageModels(c *gin.Context) {
	h, err := initHelper(c, "updateStorageModels")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.updateStorageModels()
	bodies.SetAccepted(
		c,
		"received update all integrated storage models request successfully, processing",
	)
}

func deleteStorageModel(c *gin.Context) {
	h, err := initHelper(c, "deleteStorageModel")
	if err != nil {
		log.Errorf("integrations(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	found, err := cubecos.CheckStorageModelExist(h.modelReqOpts.Driver)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if !found {
		bodies.SetNotFound(c, fmt.Errorf("storage model '%s' not found", h.modelReqOpts.Driver))
		return
	}

	h.updatePeerStorageModel()
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

	err = h.checkModelTaskUpdateReq()
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
