package tunings

import (
	"fmt"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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
			Path:    "/tunings/parameters",
			Func:    getTunings,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/tunings/specs",
			Func:    getTuningSpecs,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/parameters/:parameterName",
			Func:    patchTuning,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings",
			Func:    patchTunings,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/tunings/tasks/:taskId",
			Func:    updateTuningTask,
		},
		{
			Version: api.V1,
			Method:  http.MethodDelete,
			Path:    "/tuning/parameters/:parameterName",
			Func:    deleteTuning,
		},
		{
			Version: api.V1,
			Method:  http.MethodDelete,
			Path:    "/tunings",
			Func:    deleteTunings,
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

func patchTuning(c *gin.Context) {
	h, err := initReqHelper(c, "patchTuning")
	if err != nil {
		log.Errorf("request(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.parseTuningRequest()
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

func patchTunings(c *gin.Context) {
	tunings, err := decodeTuningsReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tunings: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	delegateTuningsReq(tunings)
	api.SetStatusOk(
		c,
		"request received and applying",
		tunings,
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

func deleteTuning(c *gin.Context) {
	h, err := initReqHelper(c, "deleteTuning")
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

	if !definition.ShouldIHandleTheTuning(tuning.Name) {
		err := fmt.Errorf("role %s is not responsible for tuning %s", definition.CurrentRole, tuning.Name)
		log.Errorf("request(%s): %s", err.Error())
		api.SetBadRequest(c, err)
		return
	}

	tuning.SetUpdating()
	addReqRecord(*tuning)
	reqQueue.Add(tuning)

	api.SetStatusOk(
		c,
		"tuning applied",
		*tuning,
	)
}

func deleteTunings(c *gin.Context) {
	tunings, err := decodeTuningsReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tunings: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	delegateTuningsReq(tunings)
	api.SetStatusOk(
		c,
		"request received and deleting",
		tunings,
	)
}
