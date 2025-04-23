package nodes

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/node"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/nodes",
			Func:    listNodes,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/nodes/:nodeName",
			Func:    getNode,
		},
	}
)

func init() {
	go streamingWatcher()
}

func listNodes(c *gin.Context) {
	h, err := initHelper(c, "listNodes")
	if err != nil {
		log.Errorf("nodes(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	resp, err := h.listNodes()
	if err != nil {
		log.Errorf("nodes(%s): failed to gen node: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchNode(h, *resp)
		return
	}

	api.SetStatusOk(
		c,
		"fetch nodes list successfully",
		resp,
	)
}

func getNode(c *gin.Context) {
	h, err := initHelper(c, "getNode")
	if err != nil {
		log.Errorf("nodes(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	node, err := h.getNode()
	if err != nil {
		log.Errorf("nodes(%s): failed to get node details: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchNode(h, *node)
		return
	}

	api.SetStatusOk(
		c,
		"fetch node successfully",
		node,
	)
}
