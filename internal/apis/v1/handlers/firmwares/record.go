package firmwares

import (
	"fmt"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	cryptossh "golang.org/x/crypto/ssh"
)

func (h *helper) isBoostrappingInProgress() bool {
	_, err := os.Stat(firmwares.BoostrappingMarker)
	return err == nil
}

func (h *helper) setBoostrappingMarker() {
	err := os.WriteFile(firmwares.BoostrappingMarker, []byte(""), 0644)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create boostrapping marker file %s(%v)", h.reqId, firmwares.BoostrappingMarker, err)
		return
	}

	controllers, err := nodes.GetNodesByRole(nodes.RoleControl)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get controller nodes(%v)", h.reqId, err)
		return
	}

	for _, controller := range controllers {
		h.moveBoostrappingMarkerToController(controller.Hostname)
	}
}

func (h *helper) moveBoostrappingMarkerToController(node string) error {
	sshAuth, err := cubecos.GenDefaultSshAuth()
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
	err = ssh.Copy(firmwares.BoostrappingMarker, firmwares.BoostrappingMarker)
	if err != nil {
		log.Errorf("firmwares(%s): failed to copy boostrapping marker to controller %s(%v)", h.reqId, node, err)
		return err
	}

	return nil
}

func (h *helper) removePreviousBoostrappingMarker() error {
	err := os.Remove(firmwares.BoostrappingMarker)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	controllers, err := nodes.GetNodesByRole(nodes.RoleControl)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get controller nodes(%v)", h.reqId, err)
		return err
	}

	for _, controller := range controllers {
		h.removeBoostrappingMarkerFromController(controller.Hostname)
	}

	return nil
}

func (h *helper) removeBoostrappingMarkerFromController(node string) error {
	sshAuth, err := cubecos.GenDefaultSshAuth()
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
	err = ssh.Run(fmt.Sprintf("rm -f %s", firmwares.BoostrappingMarker))
	if err != nil {
		log.Errorf("firmwares(%s): failed to remove boostrapping marker from controller %s(%v)", h.reqId, node, err)
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
	// have to be implemented
	return nil
}
