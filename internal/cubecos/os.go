package cubecos

import "os/exec"

func IsExpectedEmptyStdOut(err error) bool {
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return exitErr.ExitCode() == 255
}
