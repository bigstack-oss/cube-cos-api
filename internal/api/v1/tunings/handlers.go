package tunings

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/tunings"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = tunings.ReqQueue
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/tunings/specs",
			Func:    getTuningSpecs,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/tunings/parameters",
			Func:    getTunings,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/parameters/:parameterName",
			Func:    updateTuning,
		},
		{
			Version: api.V1,
			Method:  http.MethodPut,
			Path:    "/tunings/parameters/:parameterName",
			Func:    resetTuning,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/tasks/:taskId",
			Func:    updateTuningTask,
		},
	}
)

func init() {
	go streamTunings()
}

func getTunings(c *gin.Context) {
	h, err := initReqHelper(c, "getTunings")
	if err != nil {
		log.Errorf("request(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	tunings, err := h.ListTunings()
	if err != nil {
		log.Errorf("request(%s): failed to get tunings: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchTunings(h, tunings)
		return
	}

	api.SetStatusOk(
		c,
		"fetch tuning list successfully",
		tunings,
	)
}

func getTuningSpecs(c *gin.Context) {
	h, err := initReqHelper(c, "getTuningSpecs")
	if err != nil {
		log.Errorf("request(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	specs, err := h.ListTuningSpecs()
	if err != nil {
		log.Errorf("request(%s): failed to get tuning specs: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch tuning specs successfully",
		specs,
	)
}

func updateTuning(c *gin.Context) {
	h, err := initReqHelper(c, "updateTuning")
	if err != nil {
		log.Errorf("request(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.parseTuningUpdate()
	if err != nil {
		log.Errorf("request(%s): failed to parse tuning request: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkTuningPatchReq()
	if err != nil {
		log.Errorf("request(%s): failed to check tuning: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	h.delegateTuningReq()
	api.SetStatusAccepted(
		c,
		"tuning update request received",
	)
}

func resetTuning(c *gin.Context) {
	h, err := initReqHelper(c, "resetTuning")
	if err != nil {
		log.Errorf("request(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.parseTuningReset()
	if err != nil {
		log.Errorf("request(%s): failed to parse tuning req: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkTuningResetReq()
	if err != nil {
		log.Errorf("request(%s): failed to check tuning reset: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	h.delegateTuningReq()
	api.SetStatusAccepted(
		c,
		"tuning reset request received",
	)
}

func updateTuningTask(c *gin.Context) {
	h, err := initReqHelper(c, "updateTuningTask")
	if err != nil {
		log.Errorf("request(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	tuning, err := h.decodeTuningReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkTaskUpdateReq(tuning)
	if err != nil {
		log.Errorf("request(%s): failed to check tuning: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = updateTaskStatus(tuning)
	if err != nil {
		log.Errorf("request(%s): failed to update tuning status: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"tuning status updated",
		*tuning,
	)
}
