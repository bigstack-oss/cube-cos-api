package firmwares

import (
	"encoding/json"
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateToLocal() {
	reqQueue.Add(&h.reqOpts)
}

func (h *helper) delegateToPeers(list []node, upgrade *firmwares.Upgrade) {
	for _, peer := range list {
		node, err := nodes.Get(peer.Name)
		if err != nil {
			log.Warnf("firmwares(%s): failed to get node %s (%v)", h.reqId, peer.Name, err)
			continue
		}

		if node.IsLocal() {
			continue
		}

		s := status.SystemUpdateProgress{Current: "partitioning", IsProcessing: true, ProcessPercent: 30}
		err = h.installPeer(*node)
		if err != nil {
			s.Current = "failed"
			s.IsProcessing = false
			s.ProcessPercent = 0
			s.Description = err.Error()
		}

		upgrade.Progresses = append(
			upgrade.Progresses,
			firmwares.Progress{
				Host:   node.Hostname,
				Phase:  status.Partitioning,
				Status: s,
			},
		)
	}
}

func (h *helper) installPeer(node nodes.Node) error {
	reqOpts, err := h.genPeerReq(node.Hostname)
	if err != nil {
		return err
	}

	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetBody(string(reqOpts))
	resp, err := req.Execute(h.c.Request.Method, node.UpdateFirmwareUrl())
	if err != nil {
		log.Errorf("firmwares(%s): failed to update peer %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	if resp.IsError() {
		log.Errorf("firmwares(%s): has resp error from peer %s(%s)", h.reqId, node.Hostname, resp.String())
		return err
	}

	return nil
}

func (h *helper) genPeerReq(hostname string) ([]byte, error) {
	reqOpts := deepcopy.Copy(h.reqOpts).(firmwares.ReqOpts)
	req, err := json.Marshal(reqOpts)
	if err != nil {
		log.Errorf("firmwares(%s): failed to marshal firmware request for node %s(%v)", h.reqId, hostname, err)
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
