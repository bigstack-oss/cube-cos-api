package nodes

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/nodes",
			Func:    listNodes,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/nodes/:nodeName",
			Func:    getNode,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/nodes/:nodeName/ipmi",
			Func:    verifyNodeIpmi,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/nodes/:nodeName/ipmi",
			Func:    getNodeIpmi,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPut,
			Path:    "/nodes/:nodeName/ipmi",
			Func:    updateNodeIpmi,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/nodes/:nodeName/ipmi/:operation",
			Func:    ipmiOperateNode,
		},
	}
)

func init() {
	go streamingWatcher()
}

func listNodes(c *gin.Context) {
	h, err := initHelper(c, "listNodes")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	resp, err := h.listNodes()
	if err != nil {
		log.Errorf("nodes(%s): failed to list node: %v", h.reqId, err)
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
		log.Errorf("nodes(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	node, err := h.getNode()
	if err != nil {
		log.Errorf("nodes(%s): failed to get node details: %v", h.reqId, err)
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

func verifyNodeIpmi(c *gin.Context) {
	h, err := initHelper(c, "verifyNodeIpmi")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	info, err := h.verifyNodeIpmi()
	if err != nil {
		log.Errorf("nodes(%s): failed to verify node ipmi: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"verify node ipmi successfully",
		info,
	)
}

func getNodeIpmi(c *gin.Context) {
	h, err := initHelper(c, "getNodeIpmi")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	ipmi, err := h.getNodeIpmi()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch node ipmi successfully",
		ipmi,
	)
}

func updateNodeIpmi(c *gin.Context) {
	h, err := initHelper(c, "updateNodeIpmi")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.updateNodeIpmi()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"update node ipmi successfully",
		nil,
	)
}

func ipmiOperateNode(c *gin.Context) {
	h, err := initHelper(c, "ipmiOperateNode")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.ipmiOperateNode()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"ipmi power on node successfully",
		nil,
	)
}
