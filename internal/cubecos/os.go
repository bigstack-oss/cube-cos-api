package cubecos

import "os/exec"

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
