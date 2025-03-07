package triggers

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/triggers",
			Func:    getTriggers,
		},
	}
)

func getTriggers(c *gin.Context) {
	h, err := initReqHelper(c)
	if err != nil {
		logger.Errorf("trigger(%s): failed to initReqHelper: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	triggers, err := h.getTriggers()
	if err != nil {
		logger.Errorf("trigger(%s): failed to getTriggers: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetched triggers successfully",
		triggers,
	)
}
