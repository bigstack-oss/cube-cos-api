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
	reqQueue      = storages.ReqQueue
	modelReqQueue = storages.ModelReqQueue
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

func (h *helper) updateAllStorageModelsToControllers() {
	batchStorageModelReqOpts := h.initBatchStorageModelReqOpts()
	for _, modelReqOpt := range batchStorageModelReqOpts {
		h.modelReqOpts = modelReqOpt
		h.updateStorageModelToControllers()
	}
}

func (h *helper) initBatchStorageModelReqOpts() []defstorages.ModelReqOpts {
	currents, err := h.listModels()
	if err != nil {
		log.Errorf("storages(%s): failed to list current storage models(%v)", h.reqId, err)
		return nil
	}

	inited := []defstorages.ModelReqOpts{}
	for _, modelReqOpts := range h.batchModelReqOpts {
		found := false
		for _, current := range currents {
			if modelReqOpts.Driver == current.Driver {
				found = true
				break
			}
		}

		modelReqOpts.ReqId = h.reqId
		modelReqOpts.Hostname = base.Hostname
		if found {
			modelReqOpts.SetUpdating()
		} else {
			modelReqOpts.SetDeleting()
		}

		inited = append(inited, modelReqOpts)
	}

	return inited
}

func (h *helper) updateStorageModelToControllers() {
	h.updateStorageModelToLocal()
	h.updateStorageModelToPeers()
}

func (h *helper) updateStorageModelToLocal() {
	if cubecos.IsVirtualIpOwner(base.Hostname) {
		h.addReqRecord()
	}

	modelReqQueue.Add(&h.modelReqOpts)
}

func (h *helper) updateStorageModelToPeers() {
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return
	}

	nodes, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("storages(%s): failed to get peer controller nodes: %v", h.reqId, err)
		return
	}

	for _, node := range nodes {
		h.updatePeerStorageModel(node)
	}
}

func (h *helper) updatePeerStorageModel(node nodes.Node) error {
	reqOpts, err := h.genPeerStorageModelReq(node.Hostname)
	if err != nil {
		return nil
	}

	url := h.getStorageModelUrlByHandler(node)
	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetBody(string(reqOpts))
	resp, err := req.Execute(h.c.Request.Method, url)
	if err != nil {
		log.Errorf(
			"storages(%s): failed to update peer storage model %s(%v)",
			h.reqId, node.Hostname, err,
		)
		return err
	}

	if resp.IsError() {
		log.Errorf(
			"storages(%s): has resp error during updating peer storage model on node %s(%s)",
			h.reqId, node.Hostname, resp.String(),
		)
		return err
	}

	return nil
}

func (h *helper) getStorageModelUrlByHandler(node nodes.Node) string {
	switch h.handler {
	case "creaeteStorage":
		return node.PostStorageModelUrl()
	case "updateStorage":
		return node.PatchStorageModelUrl(h.modelReqOpts.Driver)
	case "deleteStorage":
		return node.DeleteStorageModelUrl(h.modelReqOpts.Driver)
	default:
		return node.PostStorageModelUrl()
	}
}

func (h *helper) genPeerStorageModelReq(hostname string) ([]byte, error) {
	modelReqOpts := deepcopy.Copy(h.modelReqOpts).(defstorages.ModelReqOpts)
	modelReqOpts.Hostname = hostname
	req, err := json.Marshal(modelReqOpts)
	if err != nil {
		log.Errorf("storages(%s): failed to marshal storage model request for node %s(%v)", h.reqId, hostname, err)
		return nil, err
	}

	return req, nil
}
