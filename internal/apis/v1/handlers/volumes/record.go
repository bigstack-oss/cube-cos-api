package volumes

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) sendChangeEvent() {
	changes.Add(volumes.Change{Id: h.imageReqOpts.Id})
}

func (h *helper) updateImageConvertionTask() error {
	defer h.sendChangeEvent()

	switch h.imageReqOpts.Status.Current {
	case status.Error, status.Completed:
		err := h.removePendingReq()
		wait.Seconds(1)
		return err
	default:
		return h.updatePendingReq()
	}
}

func (h *helper) syncImageUploadRecord() error {
	return h.mongo.UpdateOne(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{"name": h.imageReqOpts.Name, "file": h.imageReqOpts.File, "project": h.imageReqOpts.Project, "domain": h.imageReqOpts.Domain},
		bson.M{"$set": h.imageReqOpts},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) removePendingReq() error {
	return h.mongo.DeleteAll(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{
			"name":    h.imageReqOpts.Name,
			"file":    h.imageReqOpts.File,
			"project": h.imageReqOpts.Project,
			"domain":  h.imageReqOpts.Domain,
		},
	)
}

func (h *helper) updatePendingReq() error {
	return h.mongo.UpdateOne(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{
			"name":           h.imageReqOpts.Name,
			"file":           h.imageReqOpts.File,
			"project":        h.imageReqOpts.Project,
			"domain":         h.imageReqOpts.Domain,
			"status.current": status.Importing,
		},
		bson.M{"$set": h.imageReqOpts},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) hasProcessingVolumes(list []volumes.Volume) bool {
	if len(list) == 0 {
		return false
	}

	count, err := h.mongo.GetCount(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{"status.isProcessing": true},
	)
	if err != nil {
		log.Errorf("volumes(%s): failed to count processing volumes(%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) getProcessingVolumes() ([]volumes.Volume, error) {
	c, err := h.mongo.GetQueryCursor(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{"status.isProcessing": true},
	)
	if err != nil {
		log.Errorf("volumes(%s): failed to get processing images(%v)", h.reqId, err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(20))
	defer cancel()
	defer c.Close(ctx)
	return h.parseProcessingVolumes(c)
}

func (h *helper) parseProcessingVolumes(c *mongo.Cursor) ([]volumes.Volume, error) {
	list := []volumes.Volume{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(20))
	defer cancel()

	for c.Next(ctx) {
		image := images.ReqOpts{}
		err := c.Decode(&image)
		if err != nil {
			log.Errorf("volumes(%s): failed to decode processing volume(%v)", h.reqId, err)
			continue
		}

		list = append(
			list,
			h.convertImageReqOptsToVolume(image),
		)
	}

	err := c.Err()
	if err != nil {
		log.Errorf("volumes(%s): failed to iterate processing volume(%v)", h.reqId, err)
		return nil, err
	}

	return list, nil
}

func (h *helper) convertImageReqOptsToVolume(req images.ReqOpts) volumes.Volume {
	return volumes.Volume{
		Name:       req.Name,
		Type:       req.Destination,
		DiskTag:    "os disk",
		AttachedTo: "",
		Bootable:   true,
		Shared:     false,
		SizeMiB:    req.SizeMiB,
		Status: status.Volume{
			Current:        req.Status.Current,
			IsProcessing:   req.Status.IsProcessing,
			ProcessPercent: req.Status.ProcessPercent,
		},
	}
}
