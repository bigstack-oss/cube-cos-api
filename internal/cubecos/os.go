package cubecos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	defssh "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
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

	if !IsHexSuccessful(err) {
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

func IsHexSuccessful(err error) bool {
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
		log.Errorf("os: failed to create ssh helper for soft reboot(%s)(%v)", host, err)
		return err
	}

	defer ssh.Close()
	err = ssh.Run("echo YES | hex_cli -c reboot")
	if err != nil {
		log.Errorf("os: failed to soft reboot the system %s(%v)", host, err)
		return err
	}

	return nil
}

func SetResolvedInfoBySsh(host string) error {
	sshAuth, err := defssh.GenSshAuth("/root/.ssh/id_rsa")
	if err != nil {
		log.Errorf("ssh: failed to generate ssh auth for syncing remote file to %s(%v)", host, err)
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", host)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		log.Errorf("os: failed to create ssh helper for soft reboot(%s)(%v)", host, err)
		return err
	}

	defer ssh.Close()
	err = ssh.Run(fmt.Sprintf("touch %s", firmwares.ResolvedMarker))
	if err != nil {
		log.Errorf("os: failed to set the upgrade resolved marker on the system %s(%v)", host, err)
		return err
	}

	return nil
}

func RemoveFileBySsh(host, filePath string) error {
	sshAuth, err := defssh.GenSshAuth("/root/.ssh/id_rsa")
	if err != nil {
		log.Errorf("ssh: failed to generate ssh auth for syncing remote file to %s(%v)", host, err)
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", host)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		log.Errorf("os: failed to create ssh helper for file removal(%s)(%v)", host, err)
		return err
	}

	defer ssh.Close()
	err = ssh.Run(fmt.Sprintf("rm -rf %s", filePath))
	if err != nil {
		log.Errorf("os: failed to remove the file on the system %s(%v)", host, err)
		return err
	}

	return nil
}

func PowerCycleDataCenter() error {
	out, err := exec.Command("hex_sdk", "cube_cluster_power", "cycle").CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to execute the cluster power cycle cmd(%v %s)", err, string(out))
		log.Errorf("os: %v", err)
		return err
	}

	if !IsHexSuccessful(err) {
		err := fmt.Errorf("failed to rollout nodes by power cycle by hex sdk(%s)", string(out))
		log.Errorf("os: %v", err)
		return err
	}

	log.Infof("os: successfully rollout nodes by power cycle(%s)", string(out))
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

func GetCmdReturnCode(err error) int {
	segments := strings.Split(err.Error(), "exit status ")
	if len(segments) != 2 {
		return -1
	}

	var code int
	_, scanErr := fmt.Sscanf(segments[1], "%d", &code)
	if scanErr != nil {
		return -1
	}

	return code
}
