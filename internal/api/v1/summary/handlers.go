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
	dataCenter := c.Param("datacenter")

	// M2 TODO: Check if the data center is local
	if !cubecos.IsLocalDataCenter(dataCenter) {
		return
	}

	vmOverview, err := cubecos.GetVmStatusOverview()
	if err != nil {
		log.Errorf("failed to get vm status overview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":   http.StatusInternalServerError,
			"status": "internal server error",
			"msg":    "failed to get vm status overview",
		})
		return
	}

	resourceMetrics, err := cubecos.GetResourceMetrics()
	if err != nil {
		log.Errorf("failed to get resource metrics: %v", err)
		return
	}

	roleOverview, err := cubecos.GetRoleOverview()
	if err != nil {
		log.Errorf("failed to get role overview: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch data center list successfully",
		"data": gin.H{
			"vm":      vmOverview,
			"role":    roleOverview,
			"metrics": resourceMetrics,
		},
	})
}
