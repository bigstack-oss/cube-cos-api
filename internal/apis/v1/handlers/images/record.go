package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) syncUploadRecord() error {
	return h.mongo.UpdateOne(
		images.Db,
		images.ReqCollection,
		bson.M{"name": h.reqOpts.Name, "file": h.reqOpts.File, "project": h.reqOpts.Project, "domain": h.reqOpts.Domain},
		h.reqOpts,
		options.Update().SetUpsert(true),
	)
}
