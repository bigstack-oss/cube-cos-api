package firmwares

import (
	"fmt"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	defssh "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	cryptossh "golang.org/x/crypto/ssh"
)

func (h *helper) hasBootstrappingMarker() bool {
	_, err := os.Stat(firmwares.BootstrappingMarker)
	return err == nil
}

func (h *helper) syncBootstrappingMarker() {
	err := os.WriteFile(firmwares.BootstrappingMarker, []byte(""), 0644)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create bootstrapping marker file %s(%v)", h.reqId, firmwares.BootstrappingMarker, err)
		return
	}

	for _, node := range nodes.List() {
		h.moveBootstrappingMarkerToNode(node.Hostname)
	}
}

func (h *helper) moveBootstrappingMarkerToNode(node string) error {
	sshAuth, err := defssh.GenSshAuth(defssh.DefaultPrivateKey)
	if err != nil {
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", node)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		return err
	}

	defer ssh.Close()
	err = ssh.Copy(firmwares.BootstrappingMarker, firmwares.BootstrappingMarker)
	if err != nil {
		log.Errorf("firmwares(%s): failed to copy bootstrapping marker to controller %s(%v)", h.reqId, node, err)
		return err
	}

	return nil
}

func (h *helper) removePreviousBootstrappingMarker() error {
	err := os.Remove(firmwares.BootstrappingMarker)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	controllers, err := nodes.GetNodesByRole(nodes.RoleControl)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get controller nodes(%v)", h.reqId, err)
		return err
	}

	for _, controller := range controllers {
		h.removeBootstrappingMarker(controller.Hostname)
	}

	return nil
}

func (h *helper) removeBootstrappingMarker(node string) error {
	sshAuth, err := defssh.GenSshAuth(defssh.DefaultPrivateKey)
	if err != nil {
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", node)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		return err
	}

	defer ssh.Close()
	err = ssh.Run(fmt.Sprintf("rm -f %s", firmwares.BootstrappingMarker))
	if err != nil {
		log.Errorf("firmwares(%s): failed to remove bootstrapping marker from controller %s(%v)", h.reqId, node, err)
		return err
	}

	return nil
}

func (h *helper) resetBootstrappingLog(node string) error {
	sshAuth, err := defssh.GenSshAuth(defssh.DefaultPrivateKey)
	if err != nil {
		log.Errorf("firmwares(%s): failed to generate ssh auth to remove bootstrapping logs on %s (%v)", h.reqId, node, err)
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", node)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create ssh helper to remove bootstrapping logs on %s (%v)", h.reqId, node, err)
		return err
	}

	defer ssh.Close()
	err = ssh.Run(fmt.Sprintf("echo 'reset by api' > %s", firmwares.BootstrappingLog))
	if err != nil {
		log.Errorf("firmwares(%s): failed to remove bootstrapping log from node %s(%v)", h.reqId, node, err)
		return err
	}

	return nil
}

func (h *helper) setPkgAs(status string) error {
	return h.mongo.UpdateOne(
		firmwares.Db,
		firmwares.UploadCollection,
		bson.M{},
		bson.M{"$set": bson.M{"status": status}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) checkIfHasProcessingPkg() error {
	err := h.checkPkgBy(status.Uploading)
	if err != nil {
		return err
	}

	err = h.checkPkgBy(status.Verifying)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) checkPkgBy(status string) error {
	count, err := h.mongo.GetCount(
		firmwares.Db,
		firmwares.UploadCollection,
		bson.M{"status": status},
	)
	if err != nil {
		log.Errorf("firmwares(%s): failed to check %s status(%v)", h.reqId, status, err)
		return fmt.Errorf("failed to check %s status", status)
	}

	if count > 0 {
		return fmt.Errorf(
			"there is a firmware in %s status, please try again later",
			status,
		)
	}

	return nil
}

func (h *helper) clearPkgBy(status string) error {
	err := h.mongo.DeleteAll(
		firmwares.Db,
		firmwares.UploadCollection,
		bson.M{"status": status},
	)
	if err != nil {
		log.Errorf("firmwares(%s): failed to clear %s status(%v)", h.reqId, status, err)
		return err
	}

	return nil
}

func (h *helper) updateFirmwareTask() error {
	update, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		log.Errorf("firmwares: failed to get firmware upgrade progress (%v)", err)
		return err
	}

	for i, progress := range update.Progresses {
		if progress.Host != h.reqOpts.Hostname {
			continue
		}

		current := ""
		desc := ""
		switch h.reqOpts.Status.Current {
		case status.Error:
			current = status.Failed
			desc = h.reqOpts.Status.Description
		case status.WaitingReboot:
			current = status.WaitingReboot
		}

		update.Progresses[i].Phase = status.Partitioning
		update.Progresses[i].Status = status.SystemUpdateProgress{
			Current:        current,
			IsProcessing:   true,
			ProcessPercent: 50,
			Description:    desc,
		}
		break
	}

	cubecos.SetProgressDetails(update)
	return nil
}
