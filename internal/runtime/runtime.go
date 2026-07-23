package runtime

import (
	"go-micro.dev/v5/server"
)

func NewHttpServer() (*server.Server, error) {
	err := initSystemTime()
	if err != nil {
		return nil, err
	}

	err = initIdentities()
	if err != nil {
		return nil, err
	}

	err = initDependencies()
	if err != nil {
		return nil, err
	}

	printWelcomeMessages()
	return newHttpServer()
}

func printWelcomeMessages() {
	printPromptMessages()
	printLoadedConfBody()
}
