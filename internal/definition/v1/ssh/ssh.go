package ssh

import (
	"fmt"

	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
)

func GenSshAuth(idRsaPath string) (cryptossh.AuthMethod, error) {
	key, err := os.ReadFile(idRsaPath)
	if err != nil {
		log.Errorf("fixpacks: unable to read private key(%v)", err)
		return nil, err
	}

	signer, err := cryptossh.ParsePrivateKey(key)
	if err != nil {
		log.Errorf("fixpacks: unable to parse private key(%v)", err)
		return nil, err
	}

	return cryptossh.PublicKeys(signer), nil
}

func SyncRemoteFile(host, src, dst string) error {
	sshAuth, err := GenSshAuth("/root/.ssh/id_rsa")
	if err != nil {
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", host)),
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
		log.Errorf("ssh: failed to copy firmware upgrade progress to node %s(%v)", host, err)
		return err
	}

	return nil
}
