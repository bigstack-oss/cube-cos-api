package integrations

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	mongo *mongo.Helper
	http  *http.Helper

	storageReqOpts    storages.ReqOpts
	modelReqOpts      storages.ModelReqOpts
	batchModelReqOpts []storages.ModelReqOpts
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
		mongo:   mongo.GetGlobalHelper(),
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listStorages() ([]integration.Storage, error) {
	cinders, err := cubecos.ListStorages()
	if err != nil {
		log.Errorf("integrations(%s): failed to list storages (%v)", h.reqId, err)
		return nil, err
	}

	storages := h.convertToStorages(cinders)
	h.syncProcessingStorages(&storages)
	h.sortStorages(&storages)
	return storages, nil
}

func (h *helper) listVendors() ([]string, error) {
	vendors, err := cubecos.ListVendors()
	if err != nil {
		log.Errorf("integrations(%s): failed to list vendors (%v)", h.reqId, err)
		return nil, err
	}

	h.sortVendors(&vendors)
	return vendors, nil
}

func (h *helper) listModels() ([]storages.Model, error) {
	models, err := cubecos.ListModels()
	if err != nil {
		log.Errorf("integrations(%s): failed to list models (%v)", h.reqId, err)
		return nil, err
	}

	h.syncProcessingModels(&models)
	h.sortModels(&models)
	return models, nil
}

func (h *helper) verifyStorage() (*storages.VerficationResult, error) {
	err := cubecos.CreateStorage(h.storageReqOpts.CinderDetails)
	if err != nil {
		log.Errorf("integrations(%s): failed to create storage (%v)", h.reqId, err)
		return &storages.VerficationResult{
			IsCinderServiceUp:      false,
			IsTestVolumeSuccessful: false,
		}, nil
	}

	defer func() {
		err := cubecos.DeleteStorage(h.storageReqOpts.Name)
		if err != nil {
			log.Errorf("integrations(%s): failed to delete storage after verification (%v)", h.reqId, err)
		}
	}()

	result, err := cubecos.VerifyStorage(h.storageReqOpts.Name)
	if err != nil {
		log.Errorf("integrations(%s): storage verification failed (%v)", h.reqId, err)
		return result, err
	}

	return result, nil
}
