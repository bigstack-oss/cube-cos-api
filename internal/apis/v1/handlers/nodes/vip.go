package nodes

import (
	"fmt"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
)

func (h *helper) waitForVirutalIpOwnerChanged(oldOwner string) error {
	for range 600 {
		wait.Seconds(1)
		host, err := pacemaker.GetVirtualIpHost()
		if err != nil {
			log.Errorf("nodes(%s): failed to get virtual ip host(%v)", h.reqId, err)
			continue
		}

		if host == oldOwner {
			log.Infof("nodes(%s): virtual ip owner is still %s, wait for it changed", h.reqId, oldOwner)
			continue
		}

		return nil
	}

	return fmt.Errorf(
		"failed to wait for virtual ip owner changed in 10 minutes",
	)
}

func (h *helper) moveFirmwareUpgradeProgress(node string) error {
	if !h.isProgressFileExist() {
		return nil
	}

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
	err = ssh.Copy(firmwares.UpdateProgress, firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to copy firmware upgrade progress to node %s(%v)", h.reqId, node, err)
		return err
	}

	return nil
}

func (h *helper) isProgressFileExist() bool {
	_, err := os.Stat(firmwares.UpdateProgress)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	log.Errorf(
		"nodes(%s): failed to check if firmware upgrade progress file exists(%v)",
		h.reqId, err,
	)

	return false
}
