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
			Func:    setNodeIpmi,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/nodes/:nodeName/ipmi/verify",
			Func:    verifyNodeIpmi,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/nodes/:nodeName/ipmi/:operation",
			Func:    ipmiOperateNode,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/nodes/:nodeName/ipmi/disconnect",
			Func:    disconnectNodeIpmi,
		},
	}
)

func init() {
	go streamingWatcher()
}

func listNodes(c *gin.Context) {
	h, err := initHelper(c, "listNodes")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	resp, err := h.listNodes()
	if err != nil {
		log.Errorf("nodes(%s): failed to list node(%v)", h.reqId, err)
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
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	node, err := h.getNode()
	if err != nil {
		log.Errorf("nodes(%s): failed to get node details(%v)", h.reqId, err)
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
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	info, err := h.verifyNodeIpmi()
	if err != nil {
		log.Errorf("nodes(%s): failed to verify node ipmi(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.checkBoardSerialConsistency(info)
	if err != nil {
		log.Errorf("nodes(%s): board serial mismatch(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	bodies.SetOk(
		c,
		"the node ipmi is verified successfully",
		info,
	)
}

func setNodeIpmi(c *gin.Context) {
	h, err := initHelper(c, "setNodeIpmi")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	info, err := h.verifyNodeIpmi()
	if err != nil {
		log.Errorf("nodes(%s): failed to verify node ipmi(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.checkBoardSerialConsistency(info)
	if err != nil {
		log.Errorf("nodes(%s): board serial mismatch(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.setNodeIpmi()
	if err != nil {
		log.Errorf("nodes(%s): failed to set node ipmi(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"the node ipmi setting is set successfully",
		nil,
	)
}

func ipmiOperateNode(c *gin.Context) {
	h, err := initHelper(c, "ipmiOperateNode")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.checkStatusConflict()
	if err != nil {
		bodies.SetConflict(c, err)
		return
	}

	err = h.ipmiOperateNode()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the requets of ipmi operation is accepted and under processing",
	)
}

func disconnectNodeIpmi(c *gin.Context) {
	h, err := initHelper(c, "disconnectNodeIpmi")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.disconnectNodeIpmi()
	if err != nil {
		log.Errorf("nodes(%s): failed to disconnect node ipmi(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"the ipmi is successfully disconnected",
		nil,
	)
}
