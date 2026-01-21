package fixpacks

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/fixpacks"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = fixpacks.ReqQueue
)

func (h *helper) requestOperation() {
	for _, node := range nodes.List() {
		h.addReqRecord(node.Hostname)
		if nodes.IsLocal(node.Hostname) {
			reqQueue.Add(&h.reqOpts)
		}
	}
}

func (h *helper) syncFixpackToControllers(filePath string) error {
	controllers, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get peer controllers for syncing fixpack(%v)", h.reqId, err)
		return err
	}

	for _, controller := range controllers {
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
