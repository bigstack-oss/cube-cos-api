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
			Func:    getTuningSpecs,
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
			Path:    "/tunings/tasks/:taskId",
			Func:    updateTuningTask,
		},
	}
)

func init() {
	go streamingWatcher()
}

func listTunings(c *gin.Context) {
	h, err := initHelper(c, "listTunings")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	tunings, err := h.ListTunings()
	if err != nil {
		log.Errorf("tunings(%s): failed to get tunings: %v", h.reqId, err)
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

func getTuningSpecs(c *gin.Context) {
	h, err := initHelper(c, "getTuningSpecs")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	specs, err := h.ListTuningSpecs()
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
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.parseTuningUpdate()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse tuning request: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.checkTuningPatchReq()
	if err != nil {
		log.Errorf("tunings(%s): failed to check tuning: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.delegateTuningReq()
	bodies.SetAccepted(
		c,
		"tuning update request received",
	)
}

func resetTuning(c *gin.Context) {
	h, err := initHelper(c, "resetTuning")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.parseTuningReset()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse tuning req: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.checkTuningResetReq()
	if err != nil {
		log.Errorf("tunings(%s): failed to check tuning reset: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
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
		bodies.SetBadRequest(c, err)
		return
	}

	tuning, err := h.decodeTuningReq(c.Request.Body)
	if err != nil {
		log.Errorf("tunings(%s): failed to decode tuning: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.checkTaskUpdateReq(tuning)
	if err != nil {
		log.Errorf("tunings(%s): failed to check tuning: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = updateTaskStatus(tuning)
	if err != nil {
		log.Errorf("tunings(%s): failed to update tuning status: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"tuning status updated",
		*tuning,
	)
}

func enableOrDisableTuning(c *gin.Context) {
	h, err := initHelper(c, "enableOrDisableTuning")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.parseTuningEnablement()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse tuning req: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.delegateTuningToggleReq()
	bodies.SetAccepted(
		c,
		"tuning enable or disable request received",
	)
}
