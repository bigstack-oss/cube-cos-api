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
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/nodes/:nodeName/devices",
			Func:    listNodeDevices,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/nodes/:nodeName/devices",
			Func:    addNodeDevice,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/nodes/:nodeName/devices/:device",
			Func:    removeNodeDevice,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/nodes/:nodeName/devices/:device",
			Func:    updateNodeDevice,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/nodes/:nodeName/devices/:device/osds",
			Func:    updateNodeDeviceOsds,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/nodes/:nodeName/devices/tasks",
			Func:    updateDeviceTask,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/nodes/:nodeName/osds/:id/restart",
			Func:    restartNodeOsd,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/nodes/:nodeName/osds/:id",
			Func:    removeNodeOsd,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/nodes/:nodeName/osds/:id",
			Func:    updateNodeOsd,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/nodes/:nodeName/osds/tasks",
			Func:    updateOsdTask,
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
		streamData(h, *resp)
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
		streamData(h, *node)
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

func listNodeDevices(c *gin.Context) {
	h, err := initHelper(c, "listNodeDevices")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	devices, err := h.listNodeDevices()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		streamData(h, devices)
		return
	}

	bodies.SetOk(
		c,
		"fetch node devices successfully",
		devices,
	)
}

func addNodeDevice(c *gin.Context) {
	h, err := initHelper(c, "addNodeDevice")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateDeviceReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateDeviceReq()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of creating node device is accepted and under processing",
	)
}

func removeNodeDevice(c *gin.Context) {
	h, err := initHelper(c, "removeNodeDevice")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateDeviceReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateDeviceReq()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of deleting node device is accepted and under processing",
	)
}

func updateNodeDevice(c *gin.Context) {
	h, err := initHelper(c, "updateNodeDevice")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateDeviceReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateDeviceReq()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of promoting or demoting node device is accepted and under processing",
	)
}

func updateNodeDeviceOsds(c *gin.Context) {
	h, err := initHelper(c, "updateNodeDeviceOsds")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateDeviceReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateOsdReqs()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of updating node device osds is accepted and under processing",
	)
}

func restartNodeOsd(c *gin.Context) {
	h, err := initHelper(c, "restartNodeOsd")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateOsdReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateOsdReq()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of restarting node osd is accepted and under processing",
	)
}

func removeNodeOsd(c *gin.Context) {
	h, err := initHelper(c, "removeNodeOsd")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateOsdReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateOsdReq()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of removing node osd is accepted and under processing",
	)
}

func updateNodeOsd(c *gin.Context) {
	h, err := initHelper(c, "updateNodeOsd")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.validateOsdReq()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.delegateOsdReq()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"the request of patching node osd is accepted and under processing",
	)
}

func updateDeviceTask(c *gin.Context) {
	h, err := initHelper(c, "updateDeviceTask")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.updateDeviceTask()
	if err != nil {
		log.Errorf("nodes(%s): failed to update node device task(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"update node device task successfully",
		nil,
	)
}

func updateOsdTask(c *gin.Context) {
	h, err := initHelper(c, "updateOsdTask")
	if err != nil {
		log.Errorf("nodes(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.updateOsdTask()
	if err != nil {
		log.Errorf("nodes(%s): failed to update node osd task(%v)", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"update node osd task successfully",
		nil,
	)
}
