package tunings

import (
	"errors"
	"fmt"
	"net/url"

	cubeHttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateTuningReq(tuning *definition.Tuning) {
	for _, host := range tuning.Hosts {
		node, err := definition.GetNodeByHostname(host.Name)
		if err != nil {
			log.Errorf("failed to get node by hostname(%s): %s", host.Name, err.Error())
			continue
		}

		if node.IsLocal() {
			delegateToLocal(*tuning)
			continue
		}

		err = h.delegateToOtherNode(tuning, node)
		if err != nil {
			log.Errorf("failed to delegate %s to %s: %s", tuning.Name, node.Name, err.Error())
		}
	}
}

func (h *helper) delegateToOtherNode(tuning *definition.Tuning, node *definition.Node) error {
	url := node.PatchTuningUrl(*tuning)
	body := tuning.CopyAndOverrideHost(*node)
	http := cubeHttp.GetGlobalHelper()
	resp, err := http.R().SetHeader(node.GenAuthHeader()).SetBody(body).Patch(url)
	if err != nil {
		log.Errorf("failed to send tuning %s to %s: %s", tuning.Name, node.Id, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to send tuning %s to %s: %d %s", tuning.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func delegateToLocal(tuning definition.Tuning) {
	syncRecord(tuning)
	reqQueue.Add(tuning)
}

func delegateTuningsReq(tunings []definition.Tuning) {
	for _, tuning := range tunings {
		if definition.ShouldIHandleTheTuning(tuning.Name) {
			delegateToLocal(tuning)
		}

		delegateToOtherNodes(tuning)
	}
}

func delegateToOtherNodes(tuning definition.Tuning) {
	roles, found := definition.GetRolesToHandleTuning(tuning.Name)
	if !found {
		log.Warnf("no roles to handle tuning(%s)", tuning.Name)
		return
	}

	for _, role := range roles {
		nodes, err := definition.GetNodesByRole(role.Name)
		if err == nil {
			sendTuningToOtherNodes(tuning, nodes)
			continue
		}

		if errors.Is(err, cuberr.ServiceNotFound) {
			continue
		}
		log.Errorf(
			"failed to get nodes by role(%s): %s",
			role,
			err.Error(),
		)
	}
}

func sendTuningToOtherNodes(tuning definition.Tuning, nodes []*definition.Node) {
	h := cubeHttp.GetGlobalHelper()

	for _, node := range nodes {
		resp, err := h.R().SetBody(tuning).Put(genUrl(*node, tuning))
		if !resp.IsError() && err == nil {
			continue
		}

		log.Errorf(
			"failed to send tuning %s to node %s: %d %s",
			tuning.Name,
			node.Id,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}
}

func genUrl(node definition.Node, tuning definition.Tuning) string {
	u := url.URL{
		Scheme: node.Protocol,
		Host:   node.Address,
		Path:   fmt.Sprintf("/api/v1/tunings/%s", tuning.Name),
	}
	return u.String()
}
