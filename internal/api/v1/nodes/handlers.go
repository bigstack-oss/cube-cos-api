package nodes

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
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

func init() {
	go streamNodes()
}

// TODO M1: have to check why sometime take a long time to get the nodes list
// suspect the cluster-wise license fetching might be slow by hex cli
func getNodes(c *gin.Context) {
	h, err := initReqHelper(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	resp, err := h.getNodesResp()
	if err != nil {
		log.Errorf("request(%s): failed to gen node: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchNodes(h, *resp)
		return
	}

	api.SetStatusOk(
		c,
		"fetch nodes list successfully",
		resp,
	)
}
