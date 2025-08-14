package volumes

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/volumes",
			Func:    listVolumes,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/volumes.csv",
			Func:    listVolumeAsCsv,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/volumes/images",
			Func:    convertImageToVolume,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/volumes/images/tasks",
			Func:    updateImageConvertionTask,
		},
	}
)

func init() {
	go streamWatchers()
}

func listVolumes(c *gin.Context) {
	h, err := initHelper(c, "listVolumes")
	if err != nil {
		log.Errorf("volumes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	volumes, err := h.listVolumes()
	if err != nil {
		log.Errorf("volumes(%s): failed to list volumes(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		streamData(h, *volumes)
		return
	}

	bodies.SetOk(
		c,
		"fetch volumes successfully",
		volumes,
	)
}

func listVolumeAsCsv(c *gin.Context) {
	h, err := initHelper(c, "listVolumesAsCsv")
	if err != nil {
		log.Errorf("volumes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	csv, err := h.listVolumesAsCsv()
	if err != nil {
		log.Errorf("volumes(%s): failed to list volumes(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	csv.Flush()
}

func convertImageToVolume(c *gin.Context) {
	h, err := initHelper(c, "convertImageToVolume")
	if err != nil {
		log.Errorf("volumes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.validateImageConvertionValues()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.saveUploadImage()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.delegateImageConvertionReq()
	bodies.SetAccepted(
		c,
		"the request of creating volume by image is accepted and under processing",
	)
}

func updateImageConvertionTask(c *gin.Context) {
	h, err := initHelper(c, "updateImageConvertionTask")
	if err != nil {
		log.Errorf("volumes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateImageConvertionTask()
	if err != nil {
		log.Errorf("volumes(%s): failed to update volume task(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"volume task is updated successfully",
		nil,
	)
}
