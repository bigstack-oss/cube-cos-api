package nodes

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	nodes "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	node "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	log "go-micro.dev/v5/logger"
)

var (
	devReqQueue = node.DeviceReqQueue
	osdReqQueue = node.OsdReqQueue
)

func (h *helper) delegateDeviceReq() error {
	if nodes.IsLocal(h.node) {
		h.operateLocalDevice()
		return nil
	}

	return h.operateDeviceOnPeerNode()
}

func (h *helper) delegateOsdReqs() error {
	for _, osdReqOpts := range h.osdReqOptses {
		h.osdReqOpts = osdReqOpts
		if nodes.IsLocal(h.node) {
			h.operateLocalOsd()
			continue
		}

		err := h.operateOsdOnPeerNode()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *helper) delegateOsdReq() error {
	if nodes.IsLocal(h.node) {
		h.operateLocalOsd()
		return nil
	}

	return h.operateOsdOnPeerNode()
}

func (h *helper) operateLocalDevice() {
	h.upsertDeviceReqRecord()
	devReqQueue.Add(&h.deviceReqOpts)
}

func (h *helper) operateLocalOsd() {
	h.upsertOsdReqRecord()
	osdReqQueue.Add(&h.osdReqOpts)
}

func (h *helper) operateDeviceOnPeerNode() error {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node(%s) for device request(%v)", h.reqId, h.node, err)
		return err
	}

	http := http.GetGlobalHelper()
	resp, err := http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(h.deviceReqOpts).
		Execute(
			h.getMethodByHandler(),
			h.getUrlByHandler(node),
		)
	if err != nil {
		log.Errorf("nodes(%s): failed to send device req to %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	if resp.IsError() {
		err := fmt.Errorf(
			"has error response from %s device req(%d %v)",
			node.Hostname, resp.StatusCode(), string(resp.Body()),
		)
		log.Errorf("nodes(%s): %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) operateOsdOnPeerNode() error {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node(%s) for osd request(%v)", h.reqId, h.node, err)
		return err
	}

	http := http.GetGlobalHelper()
	resp, err := http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(h.osdReqOpts).
		Execute(
			h.getMethodByHandler(),
			h.getUrlByHandler(node),
		)
	if err != nil {
		log.Errorf("nodes(%s): failed to send osd req to %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	if resp.IsError() {
		err := fmt.Errorf(
			"has error response from %s osd req(%d %v)",
			node.Hostname, resp.StatusCode(), string(resp.Body()),
		)
		log.Errorf("nodes(%s): %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) getMethodByHandler() string {
	switch h.handler {
	case "addNodeDevice":
		return "POST"
	case "updateNodeDevice":
		return "PATCH"
	case "removeNodeDevice":
		return "DELETE"
	case "restartNodeOsd":
		return "POST"
	case "removeNodeOsd":
		return "DELETE"
	case "updateNodeOsd":
		return "PATCH"
	default:
		return "GET"
	}
}

func (h *helper) getUrlByHandler(node *nodes.Node) string {
	switch h.handler {
	case "addNodeDevice":
		return node.AddDeviceUrl()
	case "updateNodeDevice":
		return node.UpdateDeviceUrl(h.deviceReqOpts.Device)
	case "removeNodeDevice":
		return node.RemoveDeviceUrl(h.deviceReqOpts.Device)
	case "restartNodeOsd":
		return node.RestartOsdUrl(h.osdReqOpts.OsdId)
	case "removeNodeOsd":
		return node.RemoveOsdUrl(h.osdReqOpts.OsdId)
	case "updateNodeOsd":
		return node.PatchOsdUrl(h.osdReqOpts.OsdId)
	default:
		return node.GetDeviceUrl(h.deviceReqOpts.Device)
	}
}
