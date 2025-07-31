package images

import (
	"os"
	"path/filepath"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "importImage":
		return h.parseImportParams()
	case "listImages":
		return h.parseListParams()
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
	return nil
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

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
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
		case <-ostime.After(2 * ostime.Second):
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
	h.reqOpts.Status.ProcessPercent = precent
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

func (h *helper) parseOs(properties map[string]any) string {
	if properties == nil {
		return ""
	}

	os, ok := properties[images.CubeDefinedOs].(string)
	if !ok || os == "" {
		return ""
	}

	return os
}

func (h *helper) parseDestination(properties map[string]any) string {
	if properties == nil {
		return ""
	}

	destination, ok := properties[images.CubeDefinedDestination].(string)
	if !ok || destination == "" {
		return ""
	}

	return destination
}

func (h *helper) parseProjectName(id string) string {
	project, err := h.openstack.GetProject(id)
	if err != nil {
		log.Errorf("images: failed to get project by id(%s): %v", id, err)
		return "unknown"
	}

	return project.Name
}

func (h *helper) parseDomain(id string) string {
	project, err := h.openstack.GetProject(id)
	if err != nil {
		log.Errorf("images: failed to get project by id(%s): %v", id, err)
		return "unknown"
	}

	return project.DomainID
}

func (h *helper) parseVisibility(visibility opsimage.ImageVisibility) string {
	switch visibility {
	case "public":
		return "public"
	case "private":
		return "private"
	case "shared":
		return "shared"
	case "community":
		return "community"
	default:
		return "unknown"
	}
}

func (h *helper) parseSizeMiB(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	return bytes / (1024 * 1024)
}

func (h *helper) parseCreatedAt(createdAt ostime.Time) string {
	if createdAt.IsZero() {
		return ""
	}

	return time.LocalRFC3339(createdAt)
}

func (h *helper) parseStatus(status opsimage.ImageStatus) string {
	switch status {
	case "queued":
		return "queued"
	case "saving":
		return "saving"
	case "importing":
		return "importing"
	case "active":
		return "active"
	case "pending_delete":
		return "pending_delete"
	case "deleted":
		return "deleted"
	case "deactivated":
		return "deactivated"
	case "killed":
		return "killed"
	default:
		return "unknown"
	}
}
