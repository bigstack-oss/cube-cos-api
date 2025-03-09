package triggers

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
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
	}
)

func listTriggers(c *gin.Context) {
	h, err := initReqHelper(c, "listTriggers")
	if err != nil {
		log.Errorf("trigger(%s): failed to initReqHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	triggers, err := h.listTriggers()
	if err != nil {
		log.Errorf("trigger(%s): failed to listTriggers: %v", api.GetReqId(c), err)
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
	h, err := initReqHelper(c, "getTrigger")
	if err != nil {
		log.Errorf("trigger(%s): failed to initReqHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	trigger, err := h.getTrigger(h.getTriggerName())
	if err != nil {
		log.Errorf("trigger(%s): failed to getTrigger: %v", api.GetReqId(c), err)
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
	h, err := initReqHelper(c, "updateTrigger")
	if err != nil {
		log.Errorf("trigger(%s): failed to initReqHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	err = h.delegateTriggerReq()
	if err != nil {
		log.Errorf("trigger(%s): failed to updateTrigger: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"updated trigger successfully",
		nil,
	)
}
