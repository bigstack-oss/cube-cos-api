package cubecos

import (
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
)

func EvacuateVms(hostname string) error {
	out, err := exec.Command("hex_sdk", "os_pre_failure_host_evacuation", hostname).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v %s", err, string(out))
		log.Errorf("hexSdk: failed to execute os_pre_failure_host_evacuation on %s(%v)", hostname, err)
		return err
	}

	if !IsHexSuccessful(err) {
		return fmt.Errorf(
			"failed to evacuate vms on %s(%v %s)",
			hostname, err, string(out),
		)
	}

	return nil
}

func WaitForAllVmsEvacuated(hostname string) error {
	openstack := openstack.GetGlobalHelper()
	maxTries := 720

	for range maxTries {
		wait.Seconds(10)
		servers, err := openstack.ListServers(servers.ListOpts{Host: hostname})
		if err != nil {
			log.Errorf("cubecos: failed to list servers on %s (%v)", hostname, err)
			continue
		}

		if len(servers) == 0 {
			log.Infof("cubecos: all vms are evacuated on %s", hostname)
			return nil
		}

		log.Infof("cubecos: migrating, still %d vms on %s ....", len(servers), hostname)
	}

	return fmt.Errorf(
		"timed out waiting for all vms evacuated on %s",
		hostname,
	)
}
