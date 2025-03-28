package triggers

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/triggers"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = triggers.ReqQueue
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/triggers",
			Func:    listTriggers,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/triggers/:triggerName",
			Func:    getTrigger,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/:triggerName",
			Func:    updateTrigger,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/:triggerName/enable",
			Func:    enableOrDisableTrigger,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/tasks/:taskId",
			Func:    updateTriggerTask,
		},
	}
)

func listTriggers(c *gin.Context) {
	h, err := initHelper(c, "listTriggers")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	triggers, err := h.listTriggers()
	if err != nil {
		log.Errorf("triggers(%s): failed to listTriggers: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetched triggers successfully",
		triggers,
	)
}

func getTrigger(c *gin.Context) {
	h, err := initHelper(c, "getTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	trigger, err := h.getTrigger(h.getTriggerName())
	if err != nil {
		log.Errorf("triggers(%s): failed to getTrigger: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetched trigger successfully",
		trigger,
	)
}

func updateTrigger(c *gin.Context) {
	h, err := initHelper(c, "updateTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	h.setUpdateInfo()
	h.delegateTriggerReq()
	api.SetStatusAccepted(
		c,
		"trigger update request received",
	)
}

func enableOrDisableTrigger(c *gin.Context) {
	h, err := initHelper(c, "enableOrDisableTrigger")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.parseTriggerEnablement()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse tuning req: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	h.delegateTriggerReq()
	api.SetStatusAccepted(
		c,
		"tuning enable or disable request received",
	)
}

func updateTriggerTask(c *gin.Context) {
	h, err := initHelper(c, "updateTriggerTask")
	if err != nil {
		log.Errorf("triggers(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkTaskUpdateReq()
	if err != nil {
		log.Errorf("triggers(%s): failed to check trigger: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.updateTaskStatus()
	if err != nil {
		log.Errorf("triggers(%s): failed to update trigger status: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"trigger status updated",
		h.trigger,
	)
}
