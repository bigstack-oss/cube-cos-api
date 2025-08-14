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
			Path:    "/triggers/tasks",
			Func:    updateTriggerTask,
		},
	}
)

func listMaterials(c *gin.Context) {
	h, err := initHelper(c, "listMaterials")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
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
		bodies.SetBadRequest(c, err, nil)
		return
	}

	if h.isMaxDryRunReached() {
		err := fmt.Errorf("maximum dry run limit reached, please try again later")
		log.Errorf("triggers(%s): %v", h.reqId, err)
		bodies.SetTooManyRequests(c, err)
		return
	}

	result, err := h.verifyMaterialScript()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
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
		bodies.SetBadRequest(c, err, nil)
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
		bodies.SetBadRequest(c, err, nil)
		return
	}

	trigger, err := h.getTrigger(h.parseTriggerName())
	if err != nil {
		log.Errorf("triggers(%s): failed to getTrigger(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
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
		bodies.SetBadRequest(c, err, nil)
		return
	}

	if h.isTriggerExist(h.reqOpts.Name) {
		err := fmt.Errorf("trigger %s already exists", h.reqOpts.Name)
		bodies.SetConflict(c, err)
		return
	}

	h.updateToControllers()
	bodies.SetAccepted(
		c,
		"the requset of applying trigger is accepted successfully",
	)
}

func updateTrigger(c *gin.Context) {
	h, err := initHelper(c, "updateTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	if !h.isTriggerExist(h.reqOpts.Name) {
		err := fmt.Errorf("trigger %s does not exist", h.reqOpts.Name)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.updateToControllers()
	bodies.SetAccepted(
		c,
		"the requset of applying trigger is accepted successfully",
	)
}

func deleteTrigger(c *gin.Context) {
	h, err := initHelper(c, "deleteTrigger")
	if err != nil {
		log.Errorf("triggers(%s): failed to initHelper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	if !h.isTriggerExist(h.reqOpts.Name) {
		err := fmt.Errorf("trigger %s does not exist", h.reqOpts.Name)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.updateToControllers()
	bodies.SetAccepted(
		c,
		"trigger deletion request received",
	)
}

func enableOrDisableTrigger(c *gin.Context) {
	h, err := initHelper(c, "enableOrDisableTrigger")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	h.updateToControllers()
	bodies.SetAccepted(
		c,
		"tuning enable or disable request received",
	)
}

func updateTriggerTask(c *gin.Context) {
	h, err := initHelper(c, "updateTriggerTask")
	if err != nil {
		log.Errorf("triggers(%s): failed to init request helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkTaskUpdateReq()
	if err != nil {
		log.Errorf("triggers(%s): failed to check trigger(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
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
		nil,
	)
}
