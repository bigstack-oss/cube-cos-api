package triggers

import (
	"fmt"
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
			Func:    listMaterials,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/triggers/materials/script/verify",
			Func:    verifyMaterialScript,
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
			Method:  http.MethodPost,
			Path:    "/triggers",
			Func:    createTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/:triggerName",
			Func:    updateTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/:triggerName/enable",
			Func:    enableOrDisableTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/triggers/:triggerName",
			Func:    deleteTrigger,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/triggers/tasks/:triggerName",
			Func:    updateTriggerTask,
		},
	}
)

func listMaterials(c *gin.Context) {
	h, err := initHelper(c, "listMaterials")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	materials, err := h.listMaterials()
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

func verifyMaterialScript(c *gin.Context) {
	h, err := initHelper(c, "verifyMaterialScript")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	result, err := h.verifyMaterialScript()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	bodies.SetOk(
		c,
		"trigger script verified successfully",
		result,
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

func createTrigger(c *gin.Context) {
	h, err := initHelper(c, "createTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if h.isTriggerExist(h.applyOpts.Name) {
		bodies.SetConflict(c, fmt.Errorf("trigger %s already exists", h.applyOpts.Name))
		return
	}

	h.setCreationReq()
	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"trigger creation request received",
	)
}

func updateTrigger(c *gin.Context) {
	h, err := initHelper(c, "updateTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.setUpdateReq()
	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"the requset of applying trigger is accepted successfully",
	)
}

func deleteTrigger(c *gin.Context) {
	h, err := initHelper(c, "deleteTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.setDeletionReq()
	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"trigger deletion request received",
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
