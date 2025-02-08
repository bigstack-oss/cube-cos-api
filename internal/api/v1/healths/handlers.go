package healths

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/healths"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = healths.ReqQueue
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/healths",
			Func:    getHealthSummary,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/healths",
			Func:    checkAndRepairAllModules,
		},
		{
			Version: api.V1,
			Method:  http.MethodPatch,
			Path:    "/healths/:module",
			Func:    forceRepairModule,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/healths/:module/history",
			Func:    getModuleHealthHistory,
		},
	}
)

func init() {
	go streamHealthSummary()
}

// TODO M1: the health info will be replaced by the real data around 2025-02-10
// there're a few implementations to need to be checked with the team.
func getHealthSummary(c *gin.Context) {
	watch, err := api.ParseWatch(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	summary := genFakeHealthSummary()
	if watch {
		watchHealthSummary(c, &summary)
		return
	}

	api.SetStatusOk(
		c,
		"fetch health summary successfully",
		summary,
	)
}

func checkAndRepairAllModules(c *gin.Context) {
	err := checkEnvCondition()
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetStatusConflict(c, err)
		return
	}

	req := genCheckRepairReq()
	reqQueue.Add(req)
	api.SetStatusAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func forceRepairModule(c *gin.Context) {
	err := checkEnvCondition()
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetStatusConflict(c, err)
		return
	}

	module, err := parseModule(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	req := genForceRepairReq(*module)
	reqQueue.Add(req)
	api.SetStatusAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func getModuleHealthHistory(c *gin.Context) {
	h, err := initReqHelper(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	checkResult := h.genFakeHealthCheckResult()
	page, err := h.genPageInfo()
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch module health history successfully",
		data{
			Health: checkResult,
			Page:   page,
		},
	)
}
