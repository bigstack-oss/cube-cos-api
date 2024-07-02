package tuning

import (
	"fmt"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/controllers/v1/tuning"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/error"
	"github.com/gin-gonic/gin"
	"github.com/mohae/deepcopy"
)

var (
	reqQueue = tuning.ReqQueue
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
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.JSON(
		http.StatusOK,
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

	c.JSON(
		http.StatusOK,
		specs,
	)
}

func applyTuning(c *gin.Context) {
	tuning, err := decodeTuningReq(c.Request.Body)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":   http.StatusBadRequest,
				"status": cuberr.BadRequest,
				"msg":    err.Error(),
			},
		)
		return
	}

	if definition.ShouldCurrentRoleHandleTheTuning(tuning.Name, definition.CurrentRole) {
		delegateToCurrentNode(*tuning)
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":   http.StatusOK,
				"status": "success",
				"msg":    "tuning applied",
			},
		)
		return
	}

	c.JSON(
		http.StatusBadRequest,
		gin.H{"error": fmt.Sprintf("role %s is not responsible for tuning %s", definition.CurrentRole, tuning.Name)},
	)
}

func applyTunings(c *gin.Context) {
	tunings, err := decodeTuningsReq(c.Request.Body)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		return
	}

	setBatchPendingUpdate(tunings)
	delegateTuningsReq(tunings)

	c.JSON(
		http.StatusOK,
		gin.H{
			"code":    http.StatusOK,
			"message": "request received and applying",
		},
	)
}

func updateTuningStatus(c *gin.Context) {
	tuning, err := decodeTuningReq(c.Request.Body)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":   http.StatusBadRequest,
				"status": cuberr.BadRequest,
				"msg":    err.Error(),
			},
		)
		return
	}

	tuning.Status.SetCurrentToCompleted()
	tuning.Status.SetDesiredToUpdate()
	err = updateRecordStatus(tuning)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":   http.StatusInternalServerError,
				"status": cuberr.InternalServerError,
				"msg":    err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"code":   http.StatusOK,
			"status": "success",
			"msg":    "tuning status updated",
		},
	)
}

func deleteTuning(c *gin.Context) {
	tuning, err := decodeTuningReq(c.Request.Body)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":   http.StatusBadRequest,
				"status": cuberr.BadRequest,
				"msg":    err.Error(),
			},
		)
		return
	}

	if !definition.ShouldCurrentRoleHandleTheTuning(tuning.Name, definition.CurrentRole) {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":   http.StatusBadRequest,
				"status": cuberr.BadRequest,
				"msg":    fmt.Sprintf("role %s is not responsible for tuning %s", definition.CurrentRole, tuning.Name),
			},
		)
		return
	}

	tuning.Status.SetDesiredToDelete()
	syncTuningRecord(*tuning)
	reqQueue.Add(tuning)
	c.JSON(
		http.StatusOK,
		gin.H{
			"code":   http.StatusOK,
			"status": "success",
			"msg":    "tuning applied",
		},
	)
}

func deleteTunings(c *gin.Context) {
	tunings, err := decodeTuningsReq(c.Request.Body)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":   http.StatusBadRequest,
				"status": cuberr.BadRequest,
				"msg":    err.Error(),
			},
		)
		return
	}

	setBatchPendingDeletion(tunings)
	delegateTuningsReq(tunings)

	c.JSON(
		http.StatusOK,
		gin.H{
			"code":   http.StatusOK,
			"status": "success",
			"msg":    "request received and deleting",
		},
	)
}
