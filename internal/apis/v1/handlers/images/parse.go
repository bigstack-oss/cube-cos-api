package images

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "importImage":
		return h.parseImportParams()
	default:
		return nil
	}
}

func (h *helper) parseImportParams() error {
	err := h.c.ShouldBind(&h.reqOpts)
	if err != nil {
		log.Errorf("images(%s): failed to parse image import params(%v)", h.reqId, err)
		return err
	}

	h.reqOpts.SetUploading()
	return nil
}

func (h *helper) saveUploadImage() error {
	image, err := h.c.FormFile("image")
	if err != nil {
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	err = h.syncUploadRecord()
	if err != nil {
		log.Errorf("images(%s): failed to insert upload record for image %s(%v)", h.reqId, h.reqOpts.File, err)
		return err
	}

	doneChan := make(chan error, 1)
	go h.runProgressWatcher(doneChan)
	doneChan <- h.c.SaveUploadedFile(image, images.GlanceDir)
	close(doneChan)
	return err
}

func (h *helper) runProgressWatcher(done <-chan error) {
	for {
		select {
		case err, ok := <-done:
			if !ok {
				return
			}

			h.syncUploadResult(err)
			return
		case <-time.After(2 * time.Second):
			h.syncUploadProgress()
		}
	}
}

func (h *helper) syncUploadProgress() {
	file, err := h.c.FormFile("image")
	if err != nil {
		log.Errorf("images(%s): failed to get form file for image %s(%v)", h.reqId, h.reqOpts.File, err)
		return
	}

	totalSize := file.Size
	if totalSize <= 0 {
		log.Errorf("images(%s): total size for image %s is zero", h.reqId, h.reqOpts.File)
		return
	}

	path := filepath.Join(images.GlanceDir, h.reqOpts.File)
	upload, err := os.Stat(path)
	if err != nil {
		log.Errorf("images(%s): failed to get file info for %s(%v)", h.reqId, path, err)
		return
	}

	precent := float64(upload.Size()) / float64(totalSize) * 100
	h.reqOpts.Status.UploadProgress = precent
	err = h.mongo.UpdateOne(
		images.Db,
		images.ReqCollection,
		bson.M{"name": h.reqOpts.Name, "file": h.reqOpts.File, "project": h.reqOpts.Project, "domain": h.reqOpts.Domain},
		bson.M{"status.uploadProgress": precent},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf("images(%s): failed to update upload progress for image %s(%v)", h.reqId, h.reqOpts.File, err)
		return
	}
}

func (h *helper) syncUploadResult(err error) {
	if err != nil {
		log.Errorf("images(%s): failed to save uploaded image %s(%v)", h.reqId, h.reqOpts.File, err)
		h.reqOpts.SetError()
	} else {
		log.Infof("images(%s): successfully saved uploaded image %s", h.reqId, h.reqOpts.File)
		h.reqOpts.SetImporting()
	}

	err = h.syncUploadRecord()
	if err != nil {
		log.Errorf(
			"images(%s): failed to insert upload record for image %s(%v)",
			h.reqId, h.reqOpts.File, err,
		)
	}
}
