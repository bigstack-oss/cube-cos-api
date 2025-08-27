package cubecos

import (
	"os"

	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
)

func GenDefaultSshAuth() (cryptossh.AuthMethod, error) {
	key, err := os.ReadFile("/root/.ssh/id_rsa")
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
