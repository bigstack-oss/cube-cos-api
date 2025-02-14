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
			Method:  http.MethodGet,
			Path:    "/healths/services/:serviceType",
			Func:    getHealthHistoryOfService,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/healths/services/:serviceType/modules/:moduleType",
			Func:    getHealthHistoryOfModule,
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
			Path:    "/healths/services/:serviceType/modules/:moduleType",
			Func:    forceRepairModule,
		},
	}
)

func init() {
	go streamHealth()
}

// M1 TODO: the health info will be replaced by the real data around 2025-02-10
// there're a few implementations to need to be checked with the team.
func getHealthSummary(c *gin.Context) {
	h, err := initReqHelper(c, "getHealthSummary")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	summary := h.genFakeHealthSummary()
	if h.watch {
		watchHealth(h, &summary)
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

func getHealthHistoryOfService(c *gin.Context) {
	h, err := initReqHelper(c, "getHealthHistoryOfService")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	history := h.genFakeHealthHistoryOfService()
	if h.watch {
		watchHealth(h, history)
		return
	}

	api.SetStatusOk(
		c,
		"fetch service health history successfully",
		history,
	)
}

func getHealthHistoryOfModule(c *gin.Context) {
	h, err := initReqHelper(c, "getHealthHistoryOfModule")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	health := h.genFakeHealthHistoryOfModule()
	if h.watch {
		watchHealth(h, health)
		return
	}

	api.SetStatusOk(
		c,
		"fetch module health history successfully",
		health,
	)
}
