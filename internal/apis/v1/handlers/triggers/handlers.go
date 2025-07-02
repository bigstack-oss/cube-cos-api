package triggers

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/triggers"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = triggers.ReqQueue
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/triggers/materials",
			Func:    listTriggerMaterials,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/triggers",
			Func:    listTriggers,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/triggers/:triggerName",
			Func:    getTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/:triggerName",
			Func:    applyTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/:triggerName/enable",
			Func:    enableOrDisableTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/tasks/:triggerName",
			Func:    updateTriggerTask,
		},
	}
)

func listTriggerMaterials(c *gin.Context) {
	h, err := initHelper(c, "listTriggerMaterials")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	materials, err := h.listTriggerMaterials()
	if err != nil {
		log.Errorf("triggers(%s): failed to list trigger materials(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetched trigger materials successfully",
		materials,
	)
}

func listTriggers(c *gin.Context) {
	h, err := initHelper(c, "listTriggers")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	triggers, err := h.listTriggers()
	if err != nil {
		log.Errorf("triggers(%s): failed to listTriggers(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetched triggers successfully",
		triggers,
	)
}

func getTrigger(c *gin.Context) {
	h, err := initHelper(c, "getTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	trigger, err := h.getTrigger(h.parseTriggerName())
	if err != nil {
		log.Errorf("triggers(%s): failed to getTrigger(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetched trigger successfully",
		trigger,
	)
}

func applyTrigger(c *gin.Context) {
	h, err := initHelper(c, "applyTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.setUpdateInfo()
	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"the requset of applying trigger is accepted successfully",
	)
}

func enableOrDisableTrigger(c *gin.Context) {
	h, err := initHelper(c, "enableOrDisableTrigger")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.parseTriggerEnablement()
	if err != nil {
		log.Errorf("tunings(%s): failed to parse tuning req(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"tuning enable or disable request received",
	)
}

func updateTriggerTask(c *gin.Context) {
	h, err := initHelper(c, "updateTriggerTask")
	if err != nil {
		log.Errorf("triggers(%s): failed to init request helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.checkTaskUpdateReq()
	if err != nil {
		log.Errorf("triggers(%s): failed to check trigger(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.updateTaskStatus()
	if err != nil {
		log.Errorf("triggers(%s): failed to update trigger status(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"trigger status updated",
		h.trigger,
	)
}
