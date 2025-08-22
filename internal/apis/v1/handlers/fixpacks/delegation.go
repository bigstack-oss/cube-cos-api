package fixpacks

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/fixpacks"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
)

var (
	reqQueue = fixpacks.ReqQueue
)

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

	ssh, err := ssh.NewHelper(
		ssh.Host(node.Hostname),
		ssh.User("root"),
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
