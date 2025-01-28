package runtime

import (
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
)

func NewHttpServer() (*server.Server, error) {
	err := initNodeIdentities()
	if err != nil {
		log.Errorf("failed to init node identities: %s", err.Error())
		return nil, err
	}

	err = initDependencies()
	if err != nil {
		log.Errorf("failed to init node clis: %s", err.Error())
		return nil, err
	}

	initNodePeerSyncer()
	initNodeApiHandler()

	showPromptMessages()
	showLoadedConfBody()

	return newHttpServer()
}
