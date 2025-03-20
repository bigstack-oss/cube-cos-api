package cubecos

import (
	"os/exec"
)

func IsOvnSFlowEnabled() bool {
	_, err := exec.Command("hex_sdk", "ovn_sflow_status").Output()
	if err == nil {
		return true
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return result.ExitCode() == 0
}
