package tunings

import (
	"errors"
	"sync/atomic"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateTuningReq() {
	sent := atomic.Int32{}
	for _, host := range h.tuning.Hosts {
		node := host.GetNode()
		if node == nil {
			log.Errorf("tunings(%s): failed to get node by hostname(%s)", h.reqId, host.Name)
			continue
		}

		if node.IsLocal() {
			go h.tuneLocal(node, &sent)
			continue
		}

		if node.IsDown() {
			log.Errorf("tunings(%s): %s is down, cannot delegate %s", h.reqId, node.Hostname, h.tuning.Name)
			continue
		}

		go h.tunePeer(node, &sent)
	}

	h.waitReqsSent(&sent)
}

func (h *helper) waitReqsSent(sent *atomic.Int32) {
	for {
		if sent.Load() == int32(len(h.tuning.Hosts)) {
			log.Infof("tunings(%s): all %s tuning requests sent", h.reqId, h.tuning.Name)
			break
		}

		wait.Seconds(1)
	}
}

func (h *helper) getTuningByNameAndHosts(name string, hosts []string) (*tunings.Tuning, error) {
	tunings, err := h.listAggregatedTunings()
	if err != nil {
		log.Errorf("tunings(%s): failed to get tuning: %v", h.reqId, err)
		return nil, err
	}

	for _, tuning := range tunings {
		if tuning.Name != name {
			continue
		}

		if !tuning.IncludeHosts(hosts) {
			continue
		}

		return &tuning, nil
	}

	return nil, errors.New("tuning not found")
}

func (h *helper) tuneLocal(node *nodes.Node, sent *atomic.Int32) {
	h.updateRecord(node.Hostname)
	reqQueue.Add(&h.tuning)
	sent.Add(1)
}

func (h *helper) tunePeer(node *nodes.Node, sent *atomic.Int32) {
	defer sent.Add(1)
	resp, err := h.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(h.genTuningBodyByHandler(node)).
		Execute(
			h.genMethodByHandler(),
			h.genTuningUrlByHandler(node),
		)
	if err != nil {
		log.Errorf("tunings(%s): failed to send %s to %s: %v", h.reqId, h.tuning.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf("tunings(%s): failed to send %s to %s: %s", h.reqId, h.tuning.Name, node.Hostname, string(resp.Body()))
	}
}

func (h *helper) genTuningBodyByHandler(node *nodes.Node) any {
	switch h.handler {
	case "enableOrDisableTuning":
		return h.genTuningEnablement(node)
	case "resetTuning":
		return h.genTuningReset(node)
	default:
		return h.genTuningUpdate(node)
	}
}

func (h *helper) genMethodByHandler() string {
	switch h.handler {
	case "resetTuning":
		return "POST"
	default:
		return "PATCH"
	}
}

func (h *helper) genTuningUrlByHandler(node *nodes.Node) string {
	switch h.handler {
	case "enableOrDisableTuning":
		return node.EnableOrDisableTuningUrl(h.tuning.Name)
	case "resetTuning":
		return node.ResetTuningUrl(h.tuning.Name)
	default:
		return node.PatchTuningUrl(h.tuning.Name)
	}
}

func (h *helper) genTuningEnablement(node *nodes.Node) *tunings.Toggle {
	return &tunings.Toggle{
		Enable: h.tuning.Enabled,
		Hosts:  []string{node.Hostname},
	}
}

func (h *helper) genTuningReset(node *nodes.Node) *tunings.Reset {
	return &tunings.Reset{
		Hosts: []string{node.Hostname},
	}
}

func (h *helper) genTuningUpdate(node *nodes.Node) *tunings.Update {
	return &tunings.Update{
		Value:   h.tuning.Value,
		Enabled: h.tuning.Enabled,
		Hosts:   []string{node.Hostname},
	}
}
