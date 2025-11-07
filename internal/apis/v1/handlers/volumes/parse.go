package volumes

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
	opsvolumes "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listVolumes", "listVolumesAsCsv":
		return h.parseListParams()
	case "convertImageToVolume":
		return h.parseConvertImageParams()
	case "updateImageConvertionTask":
		return h.parseUpdateImageConvertionTask()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	err := h.parsePage()
	if err != nil {
		return err
	}

	err = h.parseWatch()
	if err != nil {
		return err
	}

	h.parseKeyword()
	return h.parseProject()
}

func (h *helper) parseConvertImageParams() error {
	h.imageReqOpts.Id = h.reqId
	err := h.syncImageReqOpts()
	if err != nil {
		return err
	}

	h.imageReqOpts.SetUploading()
	return nil
}

func (h *helper) parseUpdateImageConvertionTask() error {
	return h.c.ShouldBindJSON(&h.imageReqOpts)
}

func (h *helper) parsePage() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	return err
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = queries.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseProject() error {
	h.project = h.c.DefaultQuery("project", "")
	return nil
}

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) syncImageUploadResult(err error) {
	if err != nil {
		log.Errorf("volumes(%s): failed to save uploaded volume %s(%v)", h.reqId, h.imageReqOpts.File, err)
		h.imageReqOpts.SetError()
	} else {
		log.Infof("volumes(%s): successfully saved uploaded volume %s", h.reqId, h.imageReqOpts.File)
	}

	err = h.syncImageUploadRecord()
	if err != nil {
		log.Errorf(
			"volumes(%s): failed to insert upload record for volume %s(%v)",
			h.reqId, h.imageReqOpts.File, err,
		)
	}
}

func (h *helper) parseDiskTag(attachments []opsvolumes.Attachment) string {
	for _, attachment := range attachments {
		if attachment.Device == "/dev/vda" || attachment.Device == "/dev/sda" {
			return "os disk"
		}
	}

	return "data disk"
}

func (h *helper) parseAttachedTo(attachments []opsvolumes.Attachment) string {
	for _, attachment := range attachments {
		if attachment.ServerID == "" {
			continue
		}

		server, err := h.openstack.GetServer(attachment.ServerID)
		if err != nil {
			log.Errorf("volumes(%s): failed to get server by id %s(%v)", h.reqId, attachment.ServerID, err)
			continue
		}

		return fmt.Sprintf(
			"%s on %s",
			attachment.Device, server.Name,
		)
	}

	return ""
}

func (h *helper) parseSizeToMiB(gib int) int64 {
	if gib <= 0 {
		return 0
	}
	return int64(math.RoundDown(float64(gib*1024), 4))
}

func (h *helper) parseCreatedAt(createdAt ostime.Time) string {
	if createdAt.IsZero() {
		return ""
	}

	return time.LocalRFC3339(createdAt)
}

func (h *helper) syncImageReqOpts() error {
	h.imageReqOpts.File = h.c.DefaultQuery("file", "")
	if h.imageReqOpts.File == "" {
		return fmt.Errorf("file parameter is required")
	}

	h.imageReqOpts.Name = h.c.DefaultQuery("name", "")
	if h.imageReqOpts.Name == "" {
		return fmt.Errorf("name parameter is required")
	}

	h.imageReqOpts.Os = h.c.DefaultQuery("os", "")
	if h.imageReqOpts.Os == "" {
		return fmt.Errorf("os parameter is required")
	}

	h.imageReqOpts.Destination = h.c.DefaultQuery("destination", "")
	if h.imageReqOpts.Destination == "" {
		return fmt.Errorf("destination parameter is required")
	}

	h.imageReqOpts.Domain = h.c.DefaultQuery("domain", "")
	if h.imageReqOpts.Domain == "" {
		return fmt.Errorf("domain parameter is required")
	}

	h.imageReqOpts.Project = h.c.DefaultQuery("project", "")
	if h.imageReqOpts.Project == "" {
		return fmt.Errorf("project parameter is required")
	}

	h.imageReqOpts.SourceFromAnotherHypervisor = h.c.DefaultQuery("sourceFromAnotherHypervisor", "false") == "true"
	h.imageReqOpts.Visibility = h.c.DefaultQuery("visibility", "private")
	h.imageReqOpts.SizeMiB = int64(math.RoundDown(float64(h.c.Request.ContentLength/1024/1024), 4))
	return nil
}

func (h *helper) saveUploadImage() error {
	err := h.syncImageUploadRecord()
	if err != nil {
		log.Errorf("volumes(%s): failed to insert upload record for volume %s(%v)", h.reqId, h.imageReqOpts.File, err)
		return err
	}

	dstPath := filepath.Join(images.GlanceDir, h.imageReqOpts.File)
	done := make(chan error, 1)
	go h.runProgressWatcher(done, dstPath)
	done <- h.SaveUploadedFile(dstPath)

	close(done)
	return nil
}

func (h *helper) SaveUploadedFile(path string) error {
	out, err := os.Create(path)
	if err != nil {
		log.Errorf("volumes(%s): failed to volume file %s(%v)", h.reqId, path, err)
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, h.c.Request.Body)
	if err != nil {
		log.Errorf("volumes(%s): failed to do volume streaming copy %s(%v)", h.reqId, path, err)
		return err
	}

	return nil
}

func (h *helper) runProgressWatcher(done <-chan error, dstPath string) {
	for {
		select {
		case err, ok := <-done:
			if !ok {
				return
			}

			h.syncImageUploadResult(err)
			return
		case <-ostime.After(2 * ostime.Second):
			h.syncImageUploadProgress(dstPath)
		}
	}
}

func (h *helper) syncImageUploadProgress(dstPath string) {
	totalSize := h.c.Request.ContentLength
	if totalSize <= 0 {
		log.Errorf("volumes(%s): total size for volume %s is zero", h.reqId, h.imageReqOpts.File)
		return
	}

	upload, err := os.Stat(dstPath)
	if err != nil {
		log.Errorf("volumes(%s): failed to get file info for %s(%v)", h.reqId, dstPath, err)
		return
	}

	precent := float64(upload.Size()) / float64(totalSize) * 100
	err = h.mongo.UpdateOne(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{"name": h.imageReqOpts.Name, "file": h.imageReqOpts.File, "project": h.imageReqOpts.Project, "domain": h.imageReqOpts.Domain},
		bson.M{"$set": bson.M{"status.processPercent": math.RoundDown(precent, 2)}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf("volumes(%s): failed to update upload progress for volume %s(%v)", h.reqId, h.imageReqOpts.File, err)
		return
	}
}
