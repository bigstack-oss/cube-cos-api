package firmwares

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) setPkgAs(status string) error {
	return h.mongo.UpdateOne(
		firmwares.Db,
		firmwares.UploadCollection,
		bson.M{},
		bson.M{"$set": bson.M{"status": status}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) checkIfHasProcessingPkg() error {
	err := h.checkPkgBy(status.Uploading)
	if err != nil {
		return err
	}

	err = h.checkPkgBy(status.Verifying)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) checkPkgBy(status string) error {
	count, err := h.mongo.GetCount(
		firmwares.Db,
		firmwares.UploadCollection,
		bson.M{"status": status},
	)
	if err != nil {
		log.Errorf("firmwares(%s): failed to check %s status(%v)", h.reqId, status, err)
		return fmt.Errorf("failed to check %s status", status)
	}

	if count > 0 {
		return fmt.Errorf(
			"there is a firmware in %s status, please try again later",
			status,
		)
	}

	return nil
}

func (h *helper) clearPkgBy(status string) error {
	err := h.mongo.DeleteAll(
		firmwares.Db,
		firmwares.UploadCollection,
		bson.M{"status": status},
	)
	if err != nil {
		log.Errorf("firmwares(%s): failed to clear %s status(%v)", h.reqId, status, err)
		return err
	}

	return nil
}
