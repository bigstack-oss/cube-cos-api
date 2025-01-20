package health

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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

// TODO M1: the health info will be replaced by the real data around 2025-02-10
// there're a few implementations to need to be checked with the team.
func listHealth(c *gin.Context) {
	api.SetStatusOkResp(
		c,
		"fetch service health list successfully",
		genFakeHealthInfo(),
	)
}

func genFakeHealthInfo() cubecos.Health {
	return cubecos.Health{
		InUse: []definition.HealthInfo{
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
		Error: []definition.HealthInfo{
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
	}
}
