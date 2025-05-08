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
			Func:    genServiceHealthHistory,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/healths/services/:serviceType/modules/:moduleType",
			Func:    getModuleHealthHistory,
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
		{
			Version: api.V1,
			Method:  http.MethodDelete,
			Path:    "/healths/tasks/repairing",
			Func:    deleteCheckRepairTask,
		},
		{
			Version: api.V1,
			Method:  http.MethodDelete,
			Path:    "/healths/tasks/repairing/:moduleType",
			Func:    deleteModuleRepairTask,
		},
	}
)

func init() {
	go streamingWatcher()
}

func getHealthSummary(c *gin.Context) {
	h, err := initHelper(c, "getHealthSummary")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	summary, err := h.getHealthSummary()
	if err != nil {
		log.Errorf("healths(%s): failed to get health summary: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

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
	h, err := initHelper(c, "checkAndRepairAllModules")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkEnvCondition()
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetStatusConflict(c, err)
		return
	}

	h.requestCheckRepair()
	api.SetStatusAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func forceRepairModule(c *gin.Context) {
	h, err := initHelper(c, "forceRepairModule")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkEnvCondition()
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetStatusConflict(c, err)
		return
	}

	h.requestForceRepair()
	api.SetStatusAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func genServiceHealthHistory(c *gin.Context) {
	h, err := initHelper(c, "genServiceHealthHistory")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	history := h.genServiceHealthHistory()
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

func getModuleHealthHistory(c *gin.Context) {
	h, err := initHelper(c, "getModuleHealthHistory")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	health := h.genModuleHealthHistory()
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

func deleteCheckRepairTask(c *gin.Context) {
	h, err := initHelper(c, "deleteCheckRepairTask")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	err = h.deleteCheckRepairTask()
	if err != nil {
		log.Errorf("healths(%s): failed to delete check repair task: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"delete check repair task successfully",
		nil,
	)
}

func deleteModuleRepairTask(c *gin.Context) {
	h, err := initHelper(c, "deleteModuleRepairTask")
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	err = h.deleteModuleCheckRepairTask()
	if err != nil {
		log.Errorf("healths(%s): failed to delete module repair task: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"delete module repair task successfully",
		nil,
	)
}
