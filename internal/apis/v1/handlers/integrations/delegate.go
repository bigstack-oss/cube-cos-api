package integrations

import (
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	defstorages "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/storages"
	"github.com/go-resty/resty/v2"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue      = storages.ReqQueue
	modelReqQueue = storages.ModelReqQueue
)

func (h *helper) updateStorage() {
	h.updateLocalStorage()
	h.updatePeerStorage()
}

func (h *helper) updateLocalStorage() {
	h.addStorageReqRecord()
	reqQueue.Add(&h.storageReqOpts)
}

func (h *helper) updatePeerStorage() {
	if !base.IsHaEnabled {
		return
	}

	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return
	}

	nodes, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("storages(%s): failed to get peer controller nodes: %v", h.reqId, err)
		return
	}

	for _, node := range nodes {
		if node.IsLocal() {
			continue
		}

		h.sendStorageReqToPeer(node)
	}
}

func (h *helper) sendStorageReqToPeer(node nodes.Node) {
	req := h.genStorageReqByHandler()
	resp, err := req.Execute(
		h.c.Request.Method,
		h.getStorageUrlByHandler(node),
	)
	if err != nil {
		log.Errorf(
			"storages(%s): failed to update peer storage %s(%v)",
			h.reqId, node.Hostname, err,
		)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"storages(%s): has resp error during updating peer storage on node %s(%s)",
			h.reqId, node.Hostname, resp.String(),
		)
		return
	}
}

func (h *helper) getStorageUrlByHandler(node nodes.Node) string {
	switch h.handler {
	case "creaeteStorage":
		return node.PostStorageUrl()
	case "setStorageAsDefault":
		return node.SetDefaultStorageUrl(h.storageReqOpts.CinderDetails.Name)
	case "verifyStorage":
		return node.VerifyStorageUrl(h.storageReqOpts.CinderDetails.Name)
	case "updateStorage":
		return node.PatchStorageUrl(h.storageReqOpts.CinderDetails.Name)
	case "deleteStorage":
		return node.DeleteStorageUrl(h.storageReqOpts.CinderDetails.Name)
	default:
		return node.PostStorageUrl()
	}
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

func (h *helper) updateStorageModels() {
	batchStorageModelReqOpts := h.initBatchStorageModelReqOpts()
	for _, modelReqOpt := range batchStorageModelReqOpts {
		h.modelReqOpts = modelReqOpt
		h.updatePeerStorageModel()
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

func (h *helper) updatePeerStorageModel() {
	h.updateLocalStorageModel()
	h.updatePeerStorageModels()
}

func (h *helper) updateLocalStorageModel() {
	h.addStorageModelReqRecord()
	modelReqQueue.Add(&h.modelReqOpts)
}

func (h *helper) updatePeerStorageModels() {
	if !base.IsHaEnabled {
		return
	}

	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return
	}

	controllers, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("storages(%s): failed to get peer controller nodes(%v)", h.reqId, err)
		return
	}

	computes, err := nodes.GetComputes()
	if err != nil {
		log.Errorf("storages(%s): failed to get compute nodes(%v)", h.reqId, err)
		return
	}

	nodesToOperate := append(controllers, computes...)
	for _, node := range nodesToOperate {
		h.sendStorageModelReqToPeer(node)
	}
}

func (h *helper) sendStorageModelReqToPeer(node nodes.Node) error {
	req := h.genStorageModelReqByHandler()
	resp, err := req.Execute(
		h.c.Request.Method,
		h.genStorageModelUrlByHandler(node),
	)
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

func (h *helper) genStorageModelUrlByHandler(node nodes.Node) string {
	switch h.handler {
	case "createStorageModel":
		return node.PostStorageModelUrl()
	case "updateStorageModel":
		return node.PutStorageModelUrl(h.modelReqOpts.Driver)
	case "updateStorageModels":
		return node.PutStorageModelsUrl()
	case "deleteStorageModel":
		return node.DeleteStorageModelUrl(h.modelReqOpts.Driver)
	default:
		return node.PostStorageModelUrl()
	}
}

func (h *helper) genStorageReqByHandler() *resty.Request {
	req := h.http.R().SetHeaders(h.convertHeadersToMap(h.c.Request.Header))
	switch h.handler {
	case "deleteStorage", "verifyStorage":
		return req
	default:
		return req.SetBody(string(h.rawBody))
	}
}

func (h *helper) genStorageModelReqByHandler() *resty.Request {
	req := h.http.R().SetHeaders(h.convertHeadersToMap(h.c.Request.Header))
	switch h.handler {
	case "deleteStorageModel":
		return req
	case "updateStorageModels":
		return req.SetFile("storageModels", defstorages.TmpUploadedStorageModelList)
	default:
		return req.SetFile("storageModel", defstorages.TmpUploadedStorageModel)
	}
}
