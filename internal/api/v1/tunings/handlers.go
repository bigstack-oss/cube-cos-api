package tunings

import (
	"fmt"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/controllers/v1/tunings"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = tunings.ReqQueue
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/tunings",
			Func:    getTunings,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/tunings/specs",
			Func:    getTuningSpecs,
		},
		{
			Version: api.V1,
			Method:  http.MethodPut,
			Path:    "/tunings/:ParameterName",
			Func:    applyTuning,
		},
		{
			Version: api.V1,
			Method:  http.MethodPut,
			Path:    "/tunings",
			Func:    applyTunings,
		},
		{
			Version: api.V1,
			Method:  http.MethodPut,
			Path:    "/tunings/:ParameterName/status",
			Func:    updateTuningStatus,
		},
		{
			Version: api.V1,
			Method:  http.MethodDelete,
			Path:    "/tuning/:ParameterName",
			Func:    deleteTuning,
		},
		{
			Version: api.V1,
			Method:  http.MethodDelete,
			Path:    "/tunings",
			Func:    deleteTunings,
		},
	}
)

func getTunings(c *gin.Context) {
	tunings, err := getTuningRecords()
	if err != nil {
		log.Errorf("request(%s): failed to get tunings: %s", api.GetReqId(c), err.Error())
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	api.SetStatusOkResp(
		c,
		"fetch tunings list successfully",
		tunings,
	)
}

// @BasePath /api/v1
// @Summary	list supported tuning specifications
// @Schemes
// @Description
// @Tags		tuning specifications
// @Success	200	{array}     string	""
// @Failure	400	{string}	string	""
// @Failure	500	{string}	string	""
// @Router		/tunings/specs [get]
func getTuningSpecs(c *gin.Context) {
	specs := []definition.TuningSpec{}
	definition.GetAllTunings().Range(func(key, value interface{}) bool {
		spec := deepcopy.Copy(value).(*definition.TuningSpec)
		spec.Roles = selectRolesUsingActivityAndLabels(spec)
		specs = append(specs, *spec)
		return true
	})

	api.SetStatusOkResp(
		c,
		"fetch tuning specs successfully",
		specs,
	)
}

func applyTuning(c *gin.Context) {
	tuning, err := decodeTuningReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning: %s", api.GetReqId(c), err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	if !definition.ShouldCurrentRoleHandleTheTuning(tuning.Name, definition.CurrentRole) {
		err := fmt.Errorf("role %s is not responsible for tuning %s", definition.CurrentRole, tuning.Name)
		log.Errorf("request(%s): %s", err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	delegateToCurrentNode(*tuning)
	api.SetStatusOkResp(
		c,
		"tuning applied",
		*tuning,
	)
}

func applyTunings(c *gin.Context) {
	tunings, err := decodeTuningsReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tunings: %s", api.GetReqId(c), err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	setBatchPendingUpdate(tunings)
	delegateTuningsReq(tunings)
	api.SetStatusOkResp(
		c,
		"request received and applying",
		tunings,
	)
}

func updateTuningStatus(c *gin.Context) {
	tuning, err := decodeTuningReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning: %s", api.GetReqId(c), err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	tuning.Status.SetCurrentToCompleted()
	tuning.Status.SetDesiredToUpdate()
	err = updateRecordStatus(tuning)
	if err != nil {
		log.Errorf("request(%s): failed to update tuning status: %s", api.GetReqId(c), err.Error())
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	api.SetStatusOkResp(
		c,
		"tuning status updated",
		*tuning,
	)
}

func deleteTuning(c *gin.Context) {
	tuning, err := decodeTuningReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning: %s", api.GetReqId(c), err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	if !definition.ShouldCurrentRoleHandleTheTuning(tuning.Name, definition.CurrentRole) {
		err := fmt.Errorf("role %s is not responsible for tuning %s", definition.CurrentRole, tuning.Name)
		log.Errorf("request(%s): %s", err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	tuning.Status.SetDesiredToDelete()
	syncTuningRecord(*tuning)
	reqQueue.Add(tuning)

	api.SetStatusOkResp(
		c,
		"tuning applied",
		*tuning,
	)
}

func deleteTunings(c *gin.Context) {
	tunings, err := decodeTuningsReq(c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tunings: %s", api.GetReqId(c), err.Error())
		api.SetErrBadRequestResp(c, err)
		return
	}

	setBatchPendingDeletion(tunings)
	delegateTuningsReq(tunings)
	api.SetStatusOkResp(
		c,
		"request received and deleting",
		tunings,
	)
}
