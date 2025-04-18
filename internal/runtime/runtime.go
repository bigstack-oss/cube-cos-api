package runtime

import (
	"go-micro.dev/v5/server"
)

func NewHttpServer() (*server.Server, error) {
	err := initIdentities()
	if err != nil {
		return nil, err
	}

	err = initDependencies()
	if err != nil {
		return nil, err
	}

	registerNodePeerSyncer()
	registerNodeApiHandler()

	showPromptMessages()
	showLoadedConfBody()

	return newHttpServer()
}
