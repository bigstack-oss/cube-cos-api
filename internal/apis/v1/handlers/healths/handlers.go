package healths

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/healths"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = healths.ReqQueue
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/healths",
			Func:    getHealthSummary,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/healths/services/:serviceType",
			Func:    genServiceHealthHistory,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/healths/services/:serviceType/modules/:moduleType",
			Func:    getModuleHealthHistory,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/healths",
			Func:    checkAndRepairAllModules,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/healths/services/:serviceType/modules/:moduleType",
			Func:    forceRepairModule,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/healths/tasks/repairing",
			Func:    deleteCheckRepairTask,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/healths/tasks/repairing/:moduleType",
			Func:    deleteModuleRepairTask,
		},
	}
)

func init() {
	go streamWatchers()
}

func getHealthSummary(c *gin.Context) {
	h, err := initHelper(c, "getHealthSummary")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	summary, err := h.getHealthSummary()
	if err != nil {
		log.Errorf("healths(%s): failed to get health summary: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchHealth(h, &summary)
		return
	}

	bodies.SetOk(
		c,
		"fetch health summary successfully",
		summary,
	)
}

func checkAndRepairAllModules(c *gin.Context) {
	h, err := initHelper(c, "checkAndRepairAllModules")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkEnvCondition()
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetConflict(c, err)
		return
	}

	h.requestCheckRepair()
	bodies.SetAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func forceRepairModule(c *gin.Context) {
	h, err := initHelper(c, "forceRepairModule")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkEnvCondition()
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetConflict(c, err)
		return
	}

	h.requestForceRepair()
	bodies.SetAccepted(
		c,
		"the request of unhealthy module repair is accepted and repairing",
	)
}

func genServiceHealthHistory(c *gin.Context) {
	h, err := initHelper(c, "genServiceHealthHistory")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	history := h.genServiceHealthHistory()
	if h.watch {
		watchHealth(h, history)
		return
	}

	bodies.SetOk(
		c,
		"fetch service health history successfully",
		history,
	)
}

func getModuleHealthHistory(c *gin.Context) {
	h, err := initHelper(c, "getModuleHealthHistory")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	health := h.genModuleHealthHistory()
	if h.watch {
		watchHealth(h, health)
		return
	}

	bodies.SetOk(
		c,
		"fetch module health history successfully",
		health,
	)
}

func deleteCheckRepairTask(c *gin.Context) {
	h, err := initHelper(c, "deleteCheckRepairTask")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.deleteCheckRepairTask()
	if err != nil {
		log.Errorf("healths(%s): failed to delete check repair task: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"delete check repair task successfully",
		nil,
	)
}

func deleteModuleRepairTask(c *gin.Context) {
	h, err := initHelper(c, "deleteModuleRepairTask")
	if err != nil {
		log.Errorf("healths(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.deleteModuleCheckRepairTask()
	if err != nil {
		log.Errorf("healths(%s): failed to delete module repair task: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"delete module repair task successfully",
		nil,
	)
}
