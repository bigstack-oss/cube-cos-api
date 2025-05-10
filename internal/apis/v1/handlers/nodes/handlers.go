package nodes

import (
	"net/http"

	api "github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
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
		log.Errorf("nodes(%s): failed to init helper: %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	resp, err := h.listNodes()
	if err != nil {
		log.Errorf("nodes(%s): failed to list node: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchNode(h, *resp)
		return
	}

	bodies.SetOk(
		c,
		"fetch nodes list successfully",
		resp,
	)
}

func getNode(c *gin.Context) {
	h, err := initHelper(c, "getNode")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper: %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	node, err := h.getNode()
	if err != nil {
		log.Errorf("nodes(%s): failed to get node details: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchNode(h, *node)
		return
	}

	bodies.SetOk(
		c,
		"fetch node successfully",
		node,
	)
}
