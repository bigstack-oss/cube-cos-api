package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) updateImageTask() error {
	switch h.reqOpts.Status.Current {
	case status.Error, status.Completed:
		return h.removePendingReq()
	default:
		return h.updatePendingReq()
	}
}

func (h *helper) syncUploadRecord() error {
	return h.mongo.UpdateOne(
		images.Db,
		images.ReqCollection,
		bson.M{"name": h.reqOpts.Name, "file": h.reqOpts.File, "project": h.reqOpts.Project, "domain": h.reqOpts.Domain},
		bson.M{"$set": h.reqOpts},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) removePendingReq() error {
	return h.mongo.DeleteAll(
		images.Db,
		images.ReqCollection,
		bson.M{
			"name":    h.reqOpts.Name,
			"file":    h.reqOpts.File,
			"project": h.reqOpts.Project,
			"domain":  h.reqOpts.Domain,
		},
	)
}

func (h *helper) updatePendingReq() error {
	return h.mongo.UpdateOne(
		images.Db,
		images.ReqCollection,
		bson.M{
			"name":           h.reqOpts.Name,
			"file":           h.reqOpts.File,
			"project":        h.reqOpts.Project,
			"domain":         h.reqOpts.Domain,
			"status.current": status.Importing,
		},
		bson.M{"$set": h.reqOpts},
		options.Update().SetUpsert(true),
	)
}
