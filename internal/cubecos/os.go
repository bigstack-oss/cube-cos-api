package cubecos

import (
	"fmt"
	"os/exec"
	"strings"

	log "go-micro.dev/v5/logger"
)

func GetSystemSerial() (string, error) {
	out, err := exec.Command("hex_sdk", "license_serial_get").Output()
	if err != nil {
		log.Errorf("base: failed to get system serial: %v", err)
		return "", err
	}

	if !IsHexSdkSuccess(err) {
		return "", fmt.Errorf("failed to get system serial by hex sdk: %v", err)
	}

	serial := strings.TrimSpace(string(out))
	return serial, nil
}

func IsExpectedEmptyStdOut(err error) bool {
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return exitErr.ExitCode() == 255
}

func IsHexSdkSuccess(err error) bool {
	if err == nil {
		return true
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return exitErr.ExitCode() == 0
}
