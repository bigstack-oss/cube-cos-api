package tunings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateTuningReq() {
	for _, host := range h.tuning.Hosts {
		node := host.GetNode()
		if node == nil {
			log.Errorf("tunings(%s): failed to get node by hostname(%s)", h.reqId, host.Name)
			continue
		}

		if node.IsLocal() {
			h.tuneLocal(node)
			continue
		}

		if node.IsDown() {
			log.Errorf("tunings(%s): %s is down, cannot delegate %s", h.reqId, node.Hostname, h.tuning.Name)
			continue
		}

		err := h.tunePeer(node)
		if err != nil {
			log.Errorf("tunings(%s): failed to delegate %s to %s: %v", h.reqId, h.tuning.Name, node.Hostname, err)
		}
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

func (h *helper) tuneLocal(node *nodes.Node) {
	h.updateRecord(node.Hostname)
	reqQueue.Add(&h.tuning)
}

func (h *helper) tunePeer(node *nodes.Node) error {
	resp, err := h.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(h.genTuningBodyByHandler(node)).
		Patch(h.genTuningUrlByHandler(node))
	if err != nil {
		log.Errorf("tunings(%s): failed to send %s to %s: %v", h.reqId, h.tuning.Name, node.Id, err)
		return err
	}

	if resp.IsError() {
		log.Errorf("tunings(%s): failed to send %s to %s: %s", h.reqId, h.tuning.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func (h *helper) genTuningBodyByHandler(node *nodes.Node) any {
	switch h.handler {
	case "enableOrDisableTuning":
		return h.genTuningEnablement(node)
	default:
		return h.genTuningUpdate(node)
	}
}

func (h *helper) genTuningUrlByHandler(node *nodes.Node) string {
	switch h.handler {
	case "enableOrDisableTuning":
		return node.EnableOrDisableTuningUrl(h.tuning.Name)
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

func (h *helper) genTuningUpdate(node *nodes.Node) *tunings.Update {
	return &tunings.Update{
		Value:   h.tuning.Value,
		Enabled: h.tuning.Enabled,
		Hosts:   []string{node.Hostname},
	}
}
