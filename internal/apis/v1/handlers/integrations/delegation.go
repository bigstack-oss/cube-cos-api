package integrations

import (
	"encoding/json"
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	defstorages "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/storages"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = storages.ReqQueue
)

func (h *helper) updateStorageToControllers() {
	h.updateStorageToLocal()
	h.updateStorageToPeers()
}

func (h *helper) updateStorageToLocal() {
	if cubecos.IsVirtualIpOwner(base.Hostname) {
		h.addReqRecord()
	}

	reqQueue.Add(&h.storageReqOpts)
}

func (h *helper) updateStorageToPeers() {
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return
	}

	nodes, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("storages(%s): failed to get peer controller nodes: %v", h.reqId, err)
		return
	}

	for _, node := range nodes {
		h.updatePeerStorage(node)
	}
}

func (h *helper) updatePeerStorage(node nodes.Node) error {
	reqOpts, err := h.genPeerStorageReq(node.Hostname)
	if err != nil {
		return nil
	}

	url := h.getStorageUrlByHandler(node)
	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetBody(string(reqOpts))
	resp, err := req.Execute(h.c.Request.Method, url)
	if err != nil {
		log.Errorf(
			"storages(%s): failed to update peer storage %s(%v)",
			h.reqId, node.Hostname, err,
		)
		return err
	}

	if resp.IsError() {
		log.Errorf(
			"storages(%s): has resp error during updating peer storage on node %s(%s)",
			h.reqId, node.Hostname, resp.String(),
		)
		return err
	}

	return nil
}

func (h *helper) getStorageUrlByHandler(node nodes.Node) string {
	switch h.handler {
	case "creaeteStorage":
		return node.PostStorageUrl()
	case "updateStorage":
		return node.PatchStorageUrl(h.storageReqOpts.Name)
	case "deleteStorage":
		return node.DeleteStorageUrl(h.storageReqOpts.Name)
	default:
		return node.PostStorageUrl()
	}
}

func (h *helper) genPeerStorageReq(hostname string) ([]byte, error) {
	reqOpts := deepcopy.Copy(h.storageReqOpts).(defstorages.ReqOpts)
	reqOpts.Hostname = hostname
	req, err := json.Marshal(reqOpts)
	if err != nil {
		log.Errorf("storages(%s): failed to marshal storage request for node %s(%v)", h.reqId, hostname, err)
		return nil, err
	}

	return req, nil
}

func (h *helper) convertHeadersToMap(headers http.Header) map[string]string {
	headerMap := map[string]string{}
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	maps.Copy(headerMap, nodes.GetSecretHeaders())
	return headerMap
}
