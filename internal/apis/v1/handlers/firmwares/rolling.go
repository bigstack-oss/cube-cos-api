package firmwares

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
	defssh "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
)

var (
	rollingTriggerCount = int32(0)
)

func (h *helper) placeRollingTrigger() {
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return
	}

	if !h.reqOpts.AutoRolling {
		return
	}

	log.Infof("firmwares(%s): start to place rolling trigger", h.reqId)
	if atomic.LoadInt32(&rollingTriggerCount) > 0 {
		return
	}

	atomic.AddInt32(&rollingTriggerCount, 1)
	defer atomic.AddInt32(&rollingTriggerCount, -1)
	for {
		log.Infof("firmwares(%s): check whether all nodes are partitioned", h.reqId)
		wait.Seconds(5)
		if !h.areAllNodesPartitioned() {
			continue
		}

		log.Infof("firmwares(%s): do rolling reboot procedure", h.reqId)
		hostname, err := cubecos.GetPrimaryControllerHost()
		if err != nil {
			log.Errorf("firmwares(%s): failed to get primary controller host(%v)", err, h.reqId)
			return
		}

		err = h.setUpdateProgressOnNode(hostname, "evacuting vms on host", status.Rebooting)
		if err != nil {
			log.Errorf("firmwares(%s): failed to set rebooting progress(%v)", err, h.reqId)
			return
		}

		if cubecos.IsPrimaryController(base.Hostname) {
			h.prerebootLocal()
			return
		}

		log.Infof("firmwares(%s): current node %s is not primary controller, send prereboot procedure request to %s", base.Hostname, h.reqId, hostname)
		h.prerebootPrimaryController()
		h.waitForPrimaryControllerVmEvacuated()
		return
	}
}

func (h *helper) prerebootLocal() {
	log.Infof("firmwares: start to evacuate vms on node %s", base.Hostname)
	err := cubecos.EvacuateVms(base.Hostname)
	if err != nil {
		log.Errorf("firmwares: failed to evacuate vms (%v)", err)
		h.markNodeAsFailed(err.Error())
		return
	}

	log.Infof("firmwares: wait for all vms evacuated on node %s", base.Hostname)
	err = cubecos.WaitForAllVmsEvacuated(base.Hostname)
	if err != nil {
		log.Errorf("firmwares: failed to wait for all vms evacuated (%v)", err)
		h.markNodeAsFailed(err.Error())
		return
	}

	log.Infof("firmwares: all vms are evacuated, set bootstrapping marker and drain node %s", base.Hostname)
	h.setBoostrappingMarker()
	err = h.drainNode()
	if err != nil {
		log.Errorf("firmwares: failed to drain node (%v)", err)
		h.markNodeAsFailed(err.Error())
		return
	}

	log.Infof("firmwares: drain completed, reboot node %s", base.Hostname)
	err = h.setUpdateProgressOnNode(base.Hostname, "rebooting", status.Rebooting)
	if err != nil {
		log.Errorf("firmwares: failed to set rebooting progress(%v)", err)
		h.markNodeAsFailed(err.Error())
		return
	}

	err = cubecos.GracefulReboot()
	if err != nil {
		log.Errorf("firmwares: failed to reboot node (%v)", err)
		h.markNodeAsFailed(err.Error())
		return
	}
}

func (h *helper) areAllNodesPartitioned() bool {
	update, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		return false
	}

	for _, progress := range update.Progresses {
		log.Infof("firmware(%s): %s %s", h.reqId, progress.Host, progress.Status.Current)
	}

	for _, progress := range update.Progresses {
		if progress.Status.Current != status.WaitingReboot {
			return false
		}
	}

	return true
}

func (h *helper) markNodeAsFailed(errMsg string) {
	update, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		return
	}

	for i, progress := range update.Progresses {
		if progress.Host != base.Hostname {
			continue
		}

		update.Progresses[i].Status.Current = status.Failed
		update.Progresses[i].Status.Description = errMsg
		break
	}

	h.setProgressDetails(update)
}

func (h *helper) drainNode() error {
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return nil
	}

	err := cubecos.MoveVirtualIpOwner()
	if err != nil {
		log.Errorf("firmwares(%s): failed to move virtual ip owner(%v)", h.reqId, err)
		return err
	}

	err = h.waitForVirutalIpOwnerChanged(base.Hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to wait for virtual ip owner changed(%v)", h.reqId, err)
		return err
	}

	node, err := pacemaker.GetVirtualIpHost()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get virtual ip host(%v)", h.reqId, err)
		return err
	}

	err = h.moveFirmwareUpgradeProgress(node)
	if err != nil {
		log.Errorf("firmwares(%s): failed to move firmware upgrade progress(%v)", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) waitForVirutalIpOwnerChanged(oldOwner string) error {
	for range 600 {
		wait.Seconds(1)
		host, err := pacemaker.GetVirtualIpHost()
		if err != nil {
			log.Errorf("firmwares(%s): failed to get virtual ip host(%v)", h.reqId, err)
			continue
		}

		if host == oldOwner {
			log.Infof("firmwares(%s): virtual ip owner is still %s, wait for it changed", h.reqId, oldOwner)
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

	sshAuth, err := defssh.GenSshAuth("/root/.ssh/id_rsa")
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
		log.Errorf("firmwares(%s): failed to copy firmware upgrade progress to node %s(%v)", h.reqId, node, err)
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
		"firmwares(%s): failed to check if firmware upgrade progress file exists(%v)",
		h.reqId, err,
	)

	return false
}
