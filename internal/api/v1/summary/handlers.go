package summary

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/summary",
			Func:    getSummary,
		},
	}
)

func getSummary(c *gin.Context) {
	vmStatusOverview, err := cubecos.GetVmStatusOverview()
	if err != nil {
		log.Errorf("request(%s): failed to get vm status overview: %v", api.GetReqId(c), err)
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	resourceMetrics, err := cubecos.GetResourceMetrics()
	if err != nil {
		log.Errorf("request(%s): failed to get resource metrics: %v", api.GetReqId(c), err)
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	roleOverview, err := cubecos.GetRoleOverview()
	if err != nil {
		log.Errorf("request(%s): failed to get role overview: %v", api.GetReqId(c), err)
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	api.SetStatusOkResp(
		c,
		"fetch summary successfully",
		cubecos.Summary{
			Vm:      *vmStatusOverview,
			Role:    *roleOverview,
			Metrics: *resourceMetrics,
		},
	)
}
