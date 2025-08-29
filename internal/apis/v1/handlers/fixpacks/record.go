package fixpacks

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) syncRequestingRecord(list *[]fixpacks.Fixpack) {
	for i, fixpack := range *list {
		filter := bson.M{"version": fixpack.Version, "status.current": status.Installing}
		if h.hasInprogressUpdate(filter) {
			(*list)[i].Status.Current = status.Installing
			(*list)[i].Status.IsProcessing = true
			continue
		}

		filter["status.current"] = status.RollingBack
		if h.hasInprogressUpdate(filter) {
			(*list)[i].Status.Current = status.RollingBack
			(*list)[i].Status.IsProcessing = true
		}
	}
}

func (h *helper) hasInprogressUpdate(filter bson.M) bool {
	count, err := h.mongo.GetCount(
		fixpacks.Db,
		fixpacks.ReqCollection,
		filter,
	)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to count in-progress record(%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) addReqRecord(node string) {
	h.reqOpts.Hostname = node
	err := h.mongo.UpdateOne(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"hostname": h.reqOpts.Hostname},
		bson.M{"$set": h.reqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"fixpacks(%s): failed to add request record(%v)",
			h.reqId, err,
		)
	}
}

func (h *helper) deleteReqRecord() error {
	err := h.mongo.DeleteOne(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"hostname": h.reqOpts.Hostname},
	)
	if err == nil {
		return nil
	}

	log.Errorf(
		"fixpacks(%s): failed to delete request record(%v)",
		h.reqId, err,
	)
	return err
}

func (h *helper) markReqRecordAsFailed() error {
	err := h.mongo.UpdateOne(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"hostname": h.reqOpts.Hostname},
		h.reqOpts,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return nil
	}

	log.Errorf(
		"fixpacks(%s): failed mark req record as failed(%v)",
		h.reqId, err,
	)
	return err
}
