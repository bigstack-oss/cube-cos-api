package images

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
	case "updateImageTask":
		return h.parseUpdateTaskParams()
	default:
		return nil
	}
}

func (h *helper) parseImportParams() error {
	h.reqOpts.Id = h.reqId
	h.reqOpts.File = h.c.DefaultQuery("file", "")
	if h.reqOpts.File == "" {
		return fmt.Errorf("file parameter is required")
	}

	h.reqOpts.Name = h.c.DefaultQuery("name", "")
	if h.reqOpts.Name == "" {
		return fmt.Errorf("name parameter is required")
	}

	h.reqOpts.Os = h.c.DefaultQuery("os", "")
	if h.reqOpts.Os == "" {
		return fmt.Errorf("os parameter is required")
	}

	h.reqOpts.Destination = h.c.DefaultQuery("destination", "")
	if h.reqOpts.Destination == "" {
		return fmt.Errorf("destination parameter is required")
	}

	h.reqOpts.Domain = h.c.DefaultQuery("domain", "")
	if h.reqOpts.Domain == "" {
		return fmt.Errorf("domain parameter is required")
	}

	h.reqOpts.Project = h.c.DefaultQuery("project", "")
	if h.reqOpts.Project == "" {
		return fmt.Errorf("project parameter is required")
	}

	h.reqOpts.SourceFromAnotherHypervisor = h.c.DefaultQuery("sourceFromAnotherHypervisor", "false") == "true"
	h.reqOpts.Visibility = h.c.DefaultQuery("visibility", "")
	if h.reqOpts.Visibility == "" {
		return fmt.Errorf("visibility parameter is required")
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

func (h *helper) parseUpdateTaskParams() error {
	return h.c.ShouldBindJSON(&h.reqOpts)
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
	err := h.syncUploadRecord()
	if err != nil {
		log.Errorf("images(%s): failed to insert upload record for image %s(%v)", h.reqId, h.reqOpts.File, err)
		return err
	}

	dstPath := filepath.Join(images.GlanceDir, h.reqOpts.File)
	done := make(chan error, 1)
	go h.runProgressWatcher(done, dstPath)
	done <- h.SaveUploadedFile(dstPath)

	close(done)
	return nil
}

func (h *helper) SaveUploadedFile(path string) error {
	out, err := os.Create(path)
	if err != nil {
		log.Errorf("images(%s): failed to image file %s(%v)", h.reqId, path, err)
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, h.c.Request.Body)
	if err != nil {
		log.Errorf("images(%s): failed to do image streaming copy %s(%v)", h.reqId, path, err)
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

			h.syncUploadResult(err)
			return
		case <-ostime.After(2 * ostime.Second):
			h.syncUploadProgress(dstPath)
		}
	}
}

func (h *helper) syncUploadProgress(dstPath string) {
	totalSize := h.c.Request.ContentLength
	if totalSize <= 0 {
		log.Errorf("images(%s): total size for image %s is zero", h.reqId, h.reqOpts.File)
		return
	}

	upload, err := os.Stat(dstPath)
	if err != nil {
		log.Errorf("images(%s): failed to get file info for %s(%v)", h.reqId, dstPath, err)
		return
	}

	precent := float64(upload.Size()) / float64(totalSize) * 100
	err = h.mongo.UpdateOne(
		images.Db,
		images.ReqCollection,
		bson.M{"name": h.reqOpts.Name, "file": h.reqOpts.File, "project": h.reqOpts.Project, "domain": h.reqOpts.Domain},
		bson.M{"$set": bson.M{"status.processPercent": math.RoundDown(precent, 2)}},
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
