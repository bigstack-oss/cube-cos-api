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
			Method:  http.MethodPost,
			Path:    "/healths/:module/repair",
			Func:    repairHealth,
		},
		{
			Version: api.V1,
			Method:  http.MethodPut,
			Path:    "/healths/:module",
			Func:    updateHealth,
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

func repairHealth(c *gin.Context) {
	err := checkRepairCondition()
	if err != nil {
		log.Errorf("failed to check repair condition: %s", err.Error())
		api.SetStatusConflict(c, err)
		return
	}

	req := genRepairReq(c)
	err = applyRepairRecord(*req)
	if err != nil {
		log.Errorf("failed to apply repair record: %s", err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	reqQueue.Add(req)
	api.SetStatusAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func updateHealth(c *gin.Context) {
	health, err := parseHealthBody(c)
	if err != nil {
		log.Errorf("failed to parse health body: %s", err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = applyRepairRecord(*health)
	if err != nil {
		log.Errorf("failed to update health repair record: %s", err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"repair status updated successfully",
		nil,
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
