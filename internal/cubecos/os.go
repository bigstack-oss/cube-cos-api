package cubecos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

const (
	strictMarker = "/etc/appliance/state/strict_mode"
)

func IsInStrictMode() bool {
	_, err := os.Stat(strictMarker)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	log.Errorf("cubecos: failed to check strict mode marker file(%v)", err)
	return true
}

func GetSystemSerial() (string, error) {
	out, err := exec.Command("hex_sdk", "license_serial_get").Output()
	if err != nil {
		log.Errorf("base: failed to get system serial(%v)", err)
		return "", err
	}

	if !IsHexSdkSuccess(err) {
		return "", fmt.Errorf("failed to get system serial by hex sdk(%v)", err)
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

func GracefulReboot() error {
	node, err := nodes.Get(base.Hostname)
	if err != nil {
		log.Errorf("os: failed to get node info(%v)", err)
		return err
	}

	if node.IsVirtualIpOwner {
		MoveVirtualIpOwner()
	}

	unix.Sync()
	return Reboot()
}

func Reboot() error {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(20))
	defer cancel()

	out, err := exec.CommandContext(ctx, "bash", "-c", "echo YES | hex_cli -c reboot").CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to reboot the system(%v %s)", err, string(out))
		log.Errorf("os: %v", err)
		return err
	}

	return nil
}

func SoftRebootBySsh(host string) error {
	ssh, err := ssh.NewHelper(
		ssh.Host(host),
		ssh.User("root"),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		return err
	}

	defer ssh.Close()
	err = ssh.Run("echo YES | hex_cli -c reboot")
	if err != nil {
		return err
	}

	return nil
}

func MoveVirtualIpOwner() error {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(5))
	defer cancel()

	out, err := exec.CommandContext(ctx, "pcs", "resource", "move", "vip").CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to move virtual ip owner(%v %s)", err, string(out))
		log.Errorf("os: %v", err)
		return err
	}

	log.Infof("os: successfully moved virtual ip owner(%s)", string(out))
	return nil
}
