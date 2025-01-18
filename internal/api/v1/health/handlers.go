package health

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/health",
			Func:    listHealth,
		},
	}
)

func listHealth(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"code":   http.StatusOK,
			"status": "ok",
			"msg":    "fetch service health list successfully",
			"data": gin.H{
				"inUse": []definition.HealthInfo{
					{
						Service: "blockStorage",
						Status:  "ok",
						Modules: []definition.Module{
							{
								Name:   "ceph",
								Status: "ok",
							},
							{
								Name:   "cephMon",
								Status: "ok",
							},
						},
					},
				},
				"error": []definition.HealthInfo{
					{
						Service: "dataPipe",
						Status:  "ng",
						Modules: []definition.Module{
							{
								Name:   "zookeeper",
								Status: "ok",
								Msg:    "",
							},
							{
								Name:   "kafka",
								Status: "ng",
								Msg:    "5 topics have no coordinator",
							},
						},
					},
				},
			},
		},
	)
}
