package images

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) sendChangeEvent() {
	changes.Add(images.Change{Id: h.reqOpts.Id})
}

func (h *helper) updateImageTask() error {
	defer h.sendChangeEvent()

	switch h.reqOpts.Status.Current {
	case status.Error, status.Completed:
		err := h.removePendingReq()
		wait.Seconds(1)
		return err
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

func (h *helper) hasProcessingImages(list []images.Image) bool {
	if len(list) == 0 {
		return false
	}

	count, err := h.mongo.GetCount(
		images.Db,
		images.ReqCollection,
		bson.M{"status.isProcessing": true},
	)
	if err != nil {
		log.Errorf("images(%s): failed to count processing images(%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) getProcessingImages() ([]images.Image, error) {
	c, err := h.mongo.GetQueryCursor(
		images.Db,
		images.ReqCollection,
		bson.M{"status.isProcessing": true},
	)
	if err != nil {
		log.Errorf("images(%s): failed to get processing images(%v)", h.reqId, err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(20))
	defer cancel()
	defer c.Close(ctx)
	return h.parseProcessingImages(c)
}

func (h *helper) parseProcessingImages(c *mongo.Cursor) ([]images.Image, error) {
	list := []images.Image{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(20))
	defer cancel()

	for c.Next(ctx) {
		image := images.ReqOpts{}
		err := c.Decode(&image)
		if err != nil {
			log.Errorf("images(%s): failed to decode processing image(%v)", h.reqId, err)
			continue
		}

		list = append(
			list,
			h.convertReqOptsToImage(image),
		)
	}

	err := c.Err()
	if err != nil {
		log.Errorf("images(%s): failed to iterate processing images(%v)", h.reqId, err)
		return nil, err
	}

	return list, nil
}

func (h *helper) convertReqOptsToImage(req images.ReqOpts) images.Image {
	return images.Image{
		Id:          req.Id,
		Name:        req.Name,
		Os:          req.Os,
		Destination: req.Destination,
		Domain:      req.Domain,
		Project:     req.Project,
		Visibility:  req.Visibility,
		SizeMiB:     req.SizeMiB,
		Status: status.Image{
			Current:        req.Status.Current,
			IsProcessing:   req.Status.IsProcessing,
			ProcessPercent: req.Status.ProcessPercent,
		},
	}
}
