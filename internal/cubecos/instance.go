package cubecos

import (
	"fmt"
	"os/exec"

	log "go-micro.dev/v5/logger"
)

func EvacuateVms(hostname string) error {
	out, err := exec.Command("hex_cli", "-c", "os_pre_failure_host_evacuation", hostname).CombinedOutput()
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
