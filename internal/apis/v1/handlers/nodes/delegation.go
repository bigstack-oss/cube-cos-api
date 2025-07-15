package nodes

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	nodes "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	node "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	log "go-micro.dev/v5/logger"
)

var (
	devReqQueue = node.DeviceReqQueue
)

func (h *helper) delegateDeviceReq() error {
	if nodes.IsLocal(h.node) {
		req := h.setDeviceCreateReq()
		h.delegateToLocal(req)
		return nil
	}

	return h.createDeviceOnPeerNode()
}

func (h *helper) delegateToLocal(req nodes.DeviceReqOpts) {
	h.upsertDeviceReqRecord()
	devReqQueue.Add(req)
}

func (h *helper) createDeviceOnPeerNode() error {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node(%s) for device request(%v)", h.reqId, h.node, err)
		return err
	}

	http := http.GetGlobalHelper()
	resp, err := http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(h.deviceReqOpts).
		Post(node.CreateDeviceUrl())
	if err != nil {
		log.Errorf("nodes(%s): failed to send device creation to %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	if resp.IsError() {
		log.Errorf(
			"nodes(%s): has error response from %s device creation(%d %v)",
			h.reqId, node.Hostname, resp.StatusCode(), string(resp.Body()),
		)
		return nil
	}

	return nil
}
