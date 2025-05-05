package tunings

import (
	"encoding/json"
	"errors"

	cubeHttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateTuningReq() {
	for _, host := range h.tuning.Hosts {
		node := host.GetNode()
		if node == nil {
			log.Errorf("tuning: failed to get node by hostname(%s)", host.Name)
			continue
		}

		h.backfillTuningInfoByHandler(h.tuning)
		if node.IsLocal() {
			b, _ := json.Marshal(h.tuning)
			log.Infof("tuning %s: node %s is local: %s", h.tuning.Name, node.Hostname, string(b))
			delegateToLocal(h.tuning)
			continue
		}

		if node.IsDown() {
			log.Errorf("tuning %s: node %s is down, cannot delegate", h.tuning.Name, node.Hostname)
			continue
		}

		err := h.delegateToOtherNode(node)
		if err != nil {
			log.Errorf("tuning: failed to delegate %s to %s: %s", h.tuning.Name, node.Hostname, err.Error())
		}
	}
}

func (h *helper) delegateTuningToggleReq() {
	for _, host := range h.tuning.Hosts {
		node := host.GetNode()
		if node == nil {
			log.Errorf("tuning: failed to get node by hostname(%s)", host.Name)
			continue
		}

		h.backfillTuningInfoByHandler(h.tuning)
		if node.IsLocal() {
			delegateToLocal(h.tuning)
			continue
		}

		if node.IsDown() {
			log.Errorf("tuning %s: node %s is down, cannot delegate", h.tuning.Name, node.Hostname)
			continue
		}

		err := h.delegateToOtherNode(node)
		if err != nil {
			log.Errorf("tuning: failed to delegate %s to %s: %s", h.tuning.Name, node.Hostname, err.Error())
		}
	}
}

func (h *helper) getTuningByNameAndHost(name, host string) (*v1.Tuning, error) {
	tunings, err := cubecos.ListTunings(v1.ListTuningOptions{AllNodes: h.allNodes})
	if err != nil {
		log.Errorf("tunings(%s): failed to get tunings: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	for _, tuning := range tunings {
		if tuning.Name != name {
			continue
		}

		if !tuning.IncludeHost(host) {
			continue
		}

		return &tuning, nil
	}

	return nil, errors.New("tuning not found")
}

func (h *helper) getTuningByNameAndHosts(name string, hosts []string) (*v1.Tuning, error) {
	tunings, err := cubecos.ListTunings(v1.ListTuningOptions{AllNodes: h.allNodes})
	if err != nil {
		log.Errorf("tunings(%s): failed to get tuning: %s", api.GetReqId(h.c), err.Error())
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

func delegateToLocal(tuning v1.Tuning) {
	addReqRecord(tuning)
	reqQueue.Add(&tuning)
}

func (h *helper) backfillTuningInfoByHandler(tuning v1.Tuning) {
	switch h.handler {
	case "updateTuning":
		h.tuning.Enabled = tuning.Enabled
	case "enableOrDisableTuning":
		h.tuning.Value = tuning.Value
	}

	h.tuning.Id = h.tuning.GenerateId()
}

func (h *helper) delegateToOtherNode(node *v1.Node) error {
	http := cubeHttp.GetGlobalHelper()
	resp, err := http.R().
		SetHeaders(v1.GenNodeAuthHeaders()).
		SetBody(genTuningUpdate(h.tuning, node)).
		Patch(node.PatchTuningUrl(h.tuning))
	if err != nil {
		log.Errorf("tunings: failed to send tuning %s to %s: %s", h.tuning.Name, node.Id, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("tunings: failed to send tuning %s to %s: %s(%d)", h.tuning.Name, node.Hostname, string(resp.Body()), resp.StatusCode())
		return errors.New(string(resp.Body()))
	}

	return nil
}

func genTuningUpdate(tuning v1.Tuning, node *v1.Node) *v1.TuningUpdate {
	return &v1.TuningUpdate{
		Value:   tuning.Value,
		Enabled: tuning.Enabled,
		Hosts:   []string{node.Hostname},
	}
}
