package fixpacks

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	operator "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/fixpacks"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = operator.ReqQueue
)

// requestOperation queues the local node and, on the VIP owner, delegates
// every other target to that node's API.
func (h *helper) requestOperation() error {
	targets, err := h.resolveTargets()
	if err != nil {
		return err
	}

	isVipOwner := cubecos.IsVirtualIpOwner(base.Hostname)
	for _, host := range targets {
		if nodes.IsLocal(host) {
			h.addReqRecord(host)
			req := h.reqOpts
			req.Hostname = host
			reqQueue.Add(&req)
			continue
		}

		if !isVipOwner {
			continue
		}

		h.addReqRecord(host)
		err := h.delegateToPeer(host)
		if err != nil {
			h.reqOpts.Hostname = host
			h.reqOpts.SetFailed()
			h.reqOpts.Status.Description = err.Error()
			h.markNodeReqRecord()
		}
	}

	return nil
}

// resolveTargets returns the requested nodes, or by default the nodes
// missing the fixpack (install) or holding it as their latest (rollback).
func (h *helper) resolveTargets() ([]string, error) {
	if len(h.reqOpts.Nodes) > 0 {
		for _, host := range h.reqOpts.Nodes {
			_, err := nodes.Get(host)
			if err != nil {
				return nil, fmt.Errorf("node %s not found", host)
			}
		}

		return h.reqOpts.Nodes, nil
	}

	if h.reqOpts.Status.Desired == status.Rollbacked {
		return cubecos.ListFixpackRollbackNodes(h.reqOpts.Version)
	}

	updatables, err := h.listUpdatableNodes(h.reqOpts.Version)
	if err != nil {
		return nil, err
	}

	targets := []string{}
	for _, n := range updatables {
		if !n.Installed {
			targets = append(targets, n.Name)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("fixpack %s is already installed on every node", h.reqOpts.Version)
	}

	return targets, nil
}

func (h *helper) delegateToPeer(host string) error {
	node, err := nodes.Get(host)
	if err != nil {
		return err
	}

	if node.Status == status.Down {
		return fmt.Errorf("unable to connect node %s", host)
	}

	url := node.InstallFixpackUrl()
	method := http.MethodPatch
	if h.reqOpts.Status.Desired == status.Rollbacked {
		url = node.RollbackFixpackUrl(h.reqOpts.Version)
		method = http.MethodPost
	} else {
		err = ssh.SyncRemoteFile(host, h.reqOpts.Path, h.reqOpts.Path)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to sync %s to %s(%v)", h.reqId, h.reqOpts.Path, host, err)
			return fmt.Errorf("unable to copy the fixpack to %s", host)
		}
	}

	body, err := json.Marshal(fixpacks.ReqOpts{Version: h.reqOpts.Version, Nodes: []string{host}})
	if err != nil {
		return err
	}

	resp, err := h.http.R().
		SetHeaders(h.peerHeaders()).
		SetBody(string(body)).
		Execute(method, url)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to delegate %s to %s(%v)", h.reqId, h.reqOpts.Version, host, err)
		return fmt.Errorf("unable to reach %s", host)
	}

	if resp.IsError() {
		log.Errorf("fixpacks(%s): %s refused %s(%s)", h.reqId, host, h.reqOpts.Version, resp.String())
		return fmt.Errorf("%s refused the request: %s", host, resp.String())
	}

	return nil
}

func (h *helper) peerHeaders() map[string]string {
	headers := map[string]string{"Content-Type": "application/json"}
	for key, values := range h.c.Request.Header {
		if len(values) > 0 && key != "Content-Length" {
			headers[key] = values[0]
		}
	}

	maps.Copy(headers, nodes.GetSecretHeaders())
	return headers
}

func (h *helper) syncFixpackToControllers(filePath string) error {
	if !base.IsHaEnabled {
		return nil
	}

	controllers, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get peer controllers for syncing fixpack(%v)", h.reqId, err)
		return err
	}

	for _, controller := range controllers {
		if controller.IsLocal() {
			continue
		}

		err := ssh.SyncRemoteFile(controller.Hostname, filePath, filePath)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to sync remote file %s to %s(%v)", h.reqId, filePath, controller.Hostname, err)
			return err
		}
	}

	return nil
}

func (h *helper) removePeerFixpacks(filePath string) error {
	controllers, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get peer controllers for removing fixpack(%v)", h.reqId, err)
		return err
	}

	for _, controller := range controllers {
		err := cubecos.RemoveFileBySsh(controller.Hostname, filePath)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to remove fixpack %s on %s(%v)", h.reqId, filePath, controller.Hostname, err)
			return err
		}
	}

	return nil
}
