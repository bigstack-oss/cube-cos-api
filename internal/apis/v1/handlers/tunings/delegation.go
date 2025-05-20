package tunings

import (
	"errors"

	bshttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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
			h.tuneLocal(h.tuning)
			continue
		}

		if node.IsDown() {
			log.Errorf("tunings(%s): node %s is down, cannot delegate %s", h.reqId, node.Hostname, h.tuning.Name)
			continue
		}

		err := h.tunePeerNode(node)
		if err != nil {
			log.Errorf("tunings(%s): failed to delegate %s to %s: %v", h.reqId, h.tuning.Name, node.Hostname, err)
		}
	}
}

func (h *helper) delegateTuningToggleReq() {
	for _, host := range h.tuning.Hosts {
		node := host.GetNode()
		if node == nil {
			log.Errorf("tunings(%s): failed to get node by hostname(%s)", h.reqId, host.Name)
			continue
		}

		h.backfillTuningInfoByHandler(h.tuning)
		if node.IsLocal() {
			h.tuneLocal(h.tuning)
			continue
		}

		if node.IsDown() {
			log.Errorf("tunings(%s): node %s is down, cannot delegate %s", h.reqId, h.tuning.Name, node.Hostname)
			continue
		}

		err := h.tunePeerNode(node)
		if err != nil {
			log.Errorf("tunings(%s): failed to delegate %s to %s: %v", h.reqId, h.tuning.Name, node.Hostname, err)
		}
	}
}

func (h *helper) getTuningByNameAndHosts(name string, hosts []string) (*tunings.Tuning, error) {
	tunings, err := cubecos.ListTunings(tunings.ListOptions{AllNodes: h.allNodes})
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

func (h *helper) tuneLocal(tuning tunings.Tuning) {
	if h.isRecordRequired {
		h.addReqRecord(tuning)
	}

	reqQueue.Add(&tuning)
}

func (h *helper) backfillTuningInfoByHandler(tuning tunings.Tuning) {
	switch h.handler {
	case "updateTuning":
		h.tuning.Enabled = tuning.Enabled
	case "enableOrDisableTuning":
		h.tuning.Value = tuning.Value
	}
}

func (h *helper) tunePeerNode(node *nodes.Node) error {
	resp, err := bshttp.GetGlobalHelper().R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(genTuningUpdate(h.tuning, node)).
		Patch(node.PatchTuningUrl(h.tuning.Name))
	if err != nil {
		log.Errorf("tunings(%s): failed to send tuning %s to %s: %v", h.reqId, h.tuning.Name, node.Id)
		return err
	}

	if resp.IsError() {
		log.Errorf("tunings(%s): failed to send tuning %s to %s: %s", h.reqId, h.tuning.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func genTuningUpdate(tuning tunings.Tuning, node *nodes.Node) *tunings.Update {
	return &tunings.Update{
		Value:   tuning.Value,
		Enabled: tuning.Enabled,
		Hosts:   []string{node.Hostname},
	}
}
