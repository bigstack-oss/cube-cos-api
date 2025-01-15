package summary

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
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
		//
		return
	}

	resourceMetrics, err := cubecos.GetResourceMetrics()
	if err != nil {
		//
		return
	}

	roleOverview, err := cubecos.GetRoleOverview()
	if err != nil {
		//
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
