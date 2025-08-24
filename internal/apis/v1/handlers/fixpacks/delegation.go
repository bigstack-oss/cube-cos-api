package fixpacks

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	deffixpacks "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/fixpacks"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
)

var (
	reqQueue = fixpacks.ReqQueue
)

func (h *helper) delegateToLocal(updatables []node) {
	for _, updatable := range updatables {
		if nodes.IsLocal(updatable.Name) {
			reqQueue.Add(&h.reqOpts)
			h.addReqRecord(updatable.Name)
			return
		}
	}
}

func (h *helper) delegateToPeers(updatables []node) {
	for _, updatable := range updatables {
		node, err := nodes.Get(updatable.Name)
		if err != nil {
			log.Warnf("fixpacks(%s): failed to get node %s (%v)", h.reqId, updatable.Name, err)
			continue
		}

		if node.IsLocal() {
			continue
		}

		h.installPeer(*node)
	}
}

func (h *helper) installPeer(node nodes.Node) {
	reqOpts, err := h.genPeerReq(node.Hostname)
	if err != nil {
		return
	}

	url := h.getUrlByHandler(node)
	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetBody(string(reqOpts))
	resp, err := req.Execute(h.c.Request.Method, url)
	if err != nil {
		log.Errorf(
			"fixpacks(%s): failed to update peer fixpack %s(%v)",
			h.reqId, node.Hostname, err,
		)
	}

	if resp.IsError() {
		log.Errorf(
			"fixpacks(%s): has resp error during updating peer fixpack on node %s(%s)",
			h.reqId, node.Hostname, resp.String(),
		)
	}
}

func (h *helper) getUrlByHandler(node nodes.Node) string {
	switch h.handler {
	case "installFixpack":
		return node.PatchFixpackUrl()
	default:
		return node.PatchFixpackUrl()
	}
}

func (h *helper) genPeerReq(hostname string) ([]byte, error) {
	reqOpts := deepcopy.Copy(h.reqOpts).(deffixpacks.ReqOpts)
	req, err := json.Marshal(reqOpts)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to marshal fixpack request for node %s(%v)", h.reqId, hostname, err)
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

func (h *helper) syncFixpack(updatables []node) error {
	if len(updatables) == 0 {
		return fmt.Errorf("no nodes to install fixpack")
	}

	for _, updatable := range updatables {
		node, err := nodes.Get(updatable.Name)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to get node %s (%v)", h.reqId, updatable.Name, err)
			return err
		}

		if node.IsLocal() {
			continue
		}

		err = h.syncFixpackToPeerNode(*node)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to sync fixpack to node %s (%v)", h.reqId, node.Hostname, err)
			return err
		}
	}

	return nil
}

func (h *helper) syncFixpackToPeerNode(node nodes.Node) error {
	path, found := cubecos.GetFixpackPathByVersion(h.reqOpts.Version)
	if !found {
		err := fmt.Errorf("fixpack %s not found", h.reqOpts.Version)
		log.Errorf("fixpacks(%s): %v", h.reqId, err)
		return err
	}

	sshAuth, err := h.genSshAuth()
	if err != nil {
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", node.Hostname)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		return err
	}

	defer ssh.Close()
	err = ssh.Copy(path, path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to copy fixpack %s to node %s(%v)", h.reqId, h.reqOpts.Version, node.Hostname, err)
		return err
	}

	return nil
}

func (h *helper) genSshAuth() (cryptossh.AuthMethod, error) {
	key, err := os.ReadFile("/root/.ssh/id_rsa")
	if err != nil {
		log.Errorf("fixpacks: unable to read private key(%v)", err)
		return nil, err
	}

	signer, err := cryptossh.ParsePrivateKey(key)
	if err != nil {
		log.Errorf("fixpacks: unable to parse private key(%v)", err)
		return nil, err
	}

	return cryptossh.PublicKeys(signer), nil
}
