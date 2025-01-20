package nodes

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
			Path:    "/nodes",
			Func:    getNodes,
		},
	}
)

// TODO M1: have to check why sometime take a long time to get the nodes list
// suspect the cluster-wise license fetching might be slow by hex cli
func getNodes(c *gin.Context) {
	nodes, err := cubecos.ListNodes()
	if err != nil {
		log.Errorf("request(%s): failed to get nodes: %s", api.GetReqId(c), err.Error())
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	addLicenseInfoToNodes(c, &nodes)
	addNodeDetailsToNodes(c, &nodes)
	api.SetStatusOkResp(
		c,
		"fetch nodes list successfully",
		nodes,
	)
}
