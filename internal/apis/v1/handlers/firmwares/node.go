package firmwares

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	cryptossh "golang.org/x/crypto/ssh"
)

func (h *helper) softRebootNode(host string) error {
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
