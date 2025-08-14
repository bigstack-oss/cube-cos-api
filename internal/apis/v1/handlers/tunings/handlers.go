package tunings

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/tunings"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = tunings.ReqQueue
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/tunings/specs",
			Func:    listTuningSpecs,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/tunings/parameters",
			Func:    listTunings,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/parameters/:parameterName",
			Func:    updateTuning,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/parameters/:parameterName/enable",
			Func:    enableOrDisableTuning,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/tunings/parameters/:parameterName/reset",
			Func:    resetTuning,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/tasks",
			Func:    updateTuningTask,
		},
	}
)

func init() {
	go streamWatchers()
}

func listTunings(c *gin.Context) {
	h, err := initHelper(c, "listTunings")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	tunings, err := h.listTunings()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchTunings(h, tunings)
		return
	}

	bodies.SetOk(
		c,
		"fetch tuning list successfully",
		tunings,
	)
}

func listTuningSpecs(c *gin.Context) {
	h, err := initHelper(c, "listTuningSpecs")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	specs, err := h.listTuningSpecs()
	if err != nil {
		log.Errorf("tunings(%s): failed to get tuning specs: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch tuning specs successfully",
		specs,
	)
}

func updateTuning(c *gin.Context) {
	h, err := initHelper(c, "updateTuning")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.parseTuningUpdate()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse update request: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkTuningPatchReq()
	if err != nil {
		log.Errorf("tunings(%s): failed to check update: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.delegateTuningReq()
	bodies.SetAccepted(
		c,
		"tuning update request received",
	)
}

func enableOrDisableTuning(c *gin.Context) {
	h, err := initHelper(c, "enableOrDisableTuning")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.parseEnableValue()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse tuning req: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.delegateTuningReq()
	bodies.SetAccepted(
		c,
		"tuning enable or disable request received",
	)
}

func resetTuning(c *gin.Context) {
	h, err := initHelper(c, "resetTuning")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.parseTuningReset()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse reset req: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkTuningResetReq()
	if err != nil {
		log.Errorf("tunings(%s): failed to check reset: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.delegateTuningReq()
	bodies.SetAccepted(
		c,
		"tuning reset request received",
	)
}

func updateTuningTask(c *gin.Context) {
	h, err := initHelper(c, "updateTuningTask")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.decodeTuningReq(c.Request.Body)
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkTaskUpdateReq()
	if err != nil {
		log.Errorf("tunings(%s): failed to check task: %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateTaskStatus()
	if err != nil {
		log.Errorf("tunings(%s): failed to update task status: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"tuning status updated",
		nil,
	)
}
