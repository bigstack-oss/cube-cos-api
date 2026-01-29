package firmwares

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateToLocal() {
	reqQueue.Add(&h.reqOpts)
}

func (h *helper) updatePeerFirmware(list []node, upgrade *firmwares.Upgrade) {
	for _, peer := range list {
		h.delegateToPeer(peer.Name, upgrade)
	}
}

func (h *helper) delegateToPeer(hostname string, upgrade *firmwares.Upgrade) error {
	node, err := nodes.Get(hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get node %s (%v)", h.reqId, hostname, err)
		return err
	}

	if node.IsLocal() {
		return nil
	}

	s := status.SystemUpdateProgress{Current: status.Installing, IsProcessing: true, ProcessPercent: 30}
	defer h.syncNodeUpgradeProgress(hostname, upgrade, &s)
	if node.Status == status.Down {
		err := fmt.Errorf("UPG1001: unable to connect node %s", node.Hostname)
		s.Current = status.Failed
		s.ProcessPercent = 0
		s.IsProcessing = false
		s.Description = err.Error()
		return err
	}

	err = h.installPeer(*node)
	if err != nil {
		err := fmt.Errorf("UPG1002: unable to install firmware on the %s", node.Hostname)
		s.Current = status.Failed
		s.IsProcessing = false
		s.ProcessPercent = 0
		s.Description = err.Error()
		return err
	}

	return nil
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

func (h *helper) prerebootPrimaryController() {
	host, err := cubecos.GetPrimaryControllerHost()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get primary controller host(%v)", h.reqId, err)
		return
	}

	node, err := nodes.Get(host)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get primary controller node %s(%v)", h.reqId, host, err)
		return
	}

	resp, err := h.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		Post(node.FirmwareRollingRebootUrl())
	if err != nil {
		log.Errorf("firmwares(%s): failed to prereboot primary controller %s(%v)", h.reqId, node.Hostname, err)
		return
	}
	if !resp.IsError() {
		return
	}

	log.Errorf(
		"firmwares(%s): %v",
		h.reqId,
		fmt.Errorf(
			"resp error for firmware rolling reboot on primary controller %s: %s",
			node.Hostname,
			string(resp.Body()),
		),
	)
}
