package images

import (
	"fmt"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/images/materials",
			Func:    listImageMaterials,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/images",
			Func:    listImages,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/images.csv",
			Func:    listImageAsCsv,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/images",
			Func:    importImage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/images/:imageId",
			Func:    updateImage,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/images/tasks",
			Func:    updateImageTask,
		},
	}
)

func init() {
	go streamWatchers()
}

func listImageMaterials(c *gin.Context) {
	h, err := initHelper(c, "listMaterials")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	materials, err := h.listMaterials()
	if err != nil {
		log.Errorf("images(%s): failed to list materials(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch image materials successfully",
		materials,
	)
}

func listImages(c *gin.Context) {
	h, err := initHelper(c, "listImages")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	images, err := h.listImages()
	if err != nil {
		log.Errorf("images(%s): failed to list images(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		streamData(h, *images)
		return
	}

	bodies.SetOk(
		c,
		"fetch images successfully",
		images,
	)
}

func listImageAsCsv(c *gin.Context) {
	h, err := initHelper(c, "listImagesAsCsv")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	csv, err := h.listImagesAsCsv()
	if err != nil {
		log.Errorf("images(%s): failed to list images(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	csv.Flush()
}

func importImage(c *gin.Context) {
	h, err := initHelper(c, "importImage")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.validateValues()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.saveUploadImage()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.delegateImageReq()
	bodies.SetAccepted(
		c,
		"the request of importing image is accepted and under processing",
	)
}

func updateImage(c *gin.Context) {
	h, err := initHelper(c, "updateImage")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	if !h.isImageExist() {
		bodies.SetNotFound(c, fmt.Errorf("image %s is not found", h.reqOpts.Name))
		return
	}

	if !h.isImageOperatable() {
		bodies.SetConflict(c, fmt.Errorf("image %s is under processing, cannot be updated", h.reqOpts.Name))
		return
	}

	err = cubecos.UpdateImage(h.reqOpts.Id, h.genImageUpdateOpts())
	if err != nil {
		log.Errorf("images(%s): failed to update image(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"image is updated successfully",
		nil,
	)
}

func updateImageTask(c *gin.Context) {
	h, err := initHelper(c, "updateImageTask")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateImageTask()
	if err != nil {
		log.Errorf("images(%s): failed to update image task(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"image task is updated successfully",
		nil,
	)
}
