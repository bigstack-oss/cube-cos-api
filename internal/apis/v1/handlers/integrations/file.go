package integrations

import (
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func (h *helper) loadStorage() error {
	storage, err := h.c.FormFile("storage")
	if err != nil {
		log.Errorf("storages(%s): %v", h.reqId, err)
		return err
	}

	err = h.c.SaveUploadedFile(storage, storages.TmpUploadedStorage)
	if err != nil {
		log.Errorf("storages(%s): failed to save storage(%v)", h.reqId, err)
		return err
	}

	payload, err := os.ReadFile(storages.TmpUploadedStorage)
	if err != nil {
		log.Errorf("storages(%s): failed to read storage file(%v)", h.reqId, err)
		return err
	}

	return yaml.Unmarshal(
		payload,
		&h.storageReqOpts.CinderDetails,
	)
}

func (h *helper) loadStorageModel() error {
	list, err := h.c.FormFile("storageModel")
	if err != nil {
		log.Errorf("storages(%s): %v", h.reqId, err)
		return err
	}

	err = h.c.SaveUploadedFile(list, storages.TmpUploadedStorageModel)
	if err != nil {
		log.Errorf("storages(%s): failed to save storage model(%v)", h.reqId, err)
		return err
	}

	payload, err := os.ReadFile(storages.TmpUploadedStorageModel)
	if err != nil {
		log.Errorf("storages(%s): failed to read storage model file(%v)", h.reqId, err)
		return err
	}

	return yaml.Unmarshal(
		payload,
		&h.modelReqOpts.Model,
	)
}

func (h *helper) loadStorageModelList() error {
	list, err := h.c.FormFile("storageModels")
	if err != nil {
		log.Errorf("storages(%s): %v", h.reqId, err)
		return err
	}

	err = h.c.SaveUploadedFile(list, storages.TmpUploadedStorageModelList)
	if err != nil {
		log.Errorf("storages(%s): failed to save storage models(%v)", h.reqId, err)
		return err
	}

	// defer os.Remove(storages.TmpUploadedStorageModelList)
	payload, err := os.ReadFile(storages.TmpUploadedStorageModelList)
	if err != nil {
		log.Errorf("storages(%s): failed to read storage models file(%v)", h.reqId, err)
		return err
	}

	models := []storages.Model{}
	err = yaml.Unmarshal(payload, &models)
	if err != nil {
		return err
	}

	for _, model := range models {
		h.batchModelReqOpts = append(
			h.batchModelReqOpts,
			storages.ModelReqOpts{Model: model},
		)
	}

	return nil
}
