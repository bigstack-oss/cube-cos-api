package tuning

import (
	"errors"
	"fmt"
	"net/url"

	cubeHttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/error"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func delegateTuningsReq(tunings []definition.Tuning) {
	for _, tuning := range tunings {
		if definition.ShouldCurrentRoleHandleTheTuning(tuning.Name, definition.CurrentRole) {
			delegateToCurrentNode(tuning)
		}

		delegateToOtherNodes(tuning)
	}
}

func delegateToCurrentNode(tuning definition.Tuning) {
	syncTuningRecord(tuning)
	reqQueue.Add(tuning)
}

func delegateToOtherNodes(tuning definition.Tuning) {
	roles, found := definition.GetRolesToHandleTuning(tuning.Name)
	if !found {
		log.Warnf("no roles to handle tuning(%s)", tuning.Name)
		return
	}

	for _, role := range roles {
		nodes, err := service.GetNodesByRole(role.Name)
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

func sendTuningToOtherNodes(tuning definition.Tuning, nodes []definition.Node) {
	h := cubeHttp.GetGlobalHelper()

	for _, node := range nodes {
		resp, err := h.R().SetBody(tuning).Put(genUrl(node, tuning))
		if !resp.IsError() && err == nil {
			continue
		}

		log.Errorf(
			"failed to send tuning %s to node %s: %d %s",
			tuning.Name,
			node.ID,
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
