package volumes

import (
	"context"
	"errors"
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
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/volumes/:volumeId/move-preflight",
			Func:    getVolumeMovePreflight,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/volumes/:volumeId/move",
			Func:    moveVolume,
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

func getVolumeMovePreflight(c *gin.Context) {
	h, err := initHelper(c, "getVolumeMovePreflight")
	if err != nil {
		log.Errorf("volumes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	pf, err := h.runMovePreflight()
	if err != nil {
		log.Errorf("volumes(%s): failed to run move preflight(%v)", h.reqId, err)
		if errors.Is(err, context.DeadlineExceeded) {
			bodies.SetGatewayTimeout(c, err)
			return
		}
		bodies.SetInternalServerError(c, err)
		return
	}

	respondMovePreflight(c, pf)
}

func moveVolume(c *gin.Context) {
	h, err := initHelper(c, "moveVolume")
	if err != nil {
		log.Errorf("volumes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	md, err := h.runMoveVolume()
	if err != nil {
		log.Errorf("volumes(%s): failed to run move volume(%v)", h.reqId, err)
		if errors.Is(err, context.DeadlineExceeded) {
			// The retype may already be running on the cluster; say the move
			// could not be confirmed rather than that it failed.
			bodies.SetGatewayTimeout(c, err)
			return
		}
		bodies.SetInternalServerError(c, err)
		return
	}

	respondMoveDispatch(c, md)
}

// respondMovePreflight maps a preflight result to its HTTP status and returns
// the full result, including every blocker, as the response data.
func respondMovePreflight(c *gin.Context, pf *cubecos.MovePreflight) {
	status := cubecos.MovePreflightStatus(pf.Code)
	switch status {
	case http.StatusOK:
		bodies.SetOk(c, pf.Reason, pf)
	case http.StatusBadRequest:
		bodies.SetBadRequest(c, errors.New(pf.Reason), pf)
	default:
		bodies.SetConflictWithData(c, errors.New(pf.Reason), pf)
	}
}

// respondMoveDispatch maps a cinder_move_volume result to its HTTP status. A
// refusal maps exactly as the preflight does and carries every blocker; a
// dispatch failure (E_DISPATCH_FAILED) is a server-side failure, not a client
// mistake; a successful dispatch stays 202 Accepted.
func respondMoveDispatch(c *gin.Context, md *cubecos.MoveDispatch) {
	status := cubecos.MoveDispatchStatus(md.Code)
	switch status {
	case http.StatusAccepted:
		bodies.SetAccepted(
			c,
			"the volume move request is accepted and dispatching",
		)
	case http.StatusBadRequest:
		bodies.SetBadRequest(c, errors.New(md.Reason), md)
	case http.StatusInternalServerError:
		bodies.SetInternalServerError(c, errors.New(md.Reason))
	default:
		bodies.SetConflictWithData(c, errors.New(md.Reason), md)
	}
}
