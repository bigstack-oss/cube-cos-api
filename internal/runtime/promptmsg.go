package runtime

import (
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	log "go-micro.dev/v5/logger"
)

func printPromptMessages() {
	log.Info("")
	log.Info(` _____       _              _____ ____   _____             _____ _____ `)
	log.Info(`/ ____|     | |            / ____/ __ \ / ____|      /\   |  __ \_   _|`)
	log.Info(`| |    _   _| |__   ___   | |   | |  | | (___       /  \  | |__) || |  `)
	log.Info(`| |   | | | | '_ \ / _ \  | |   | |  | |\___ \     / /\ \ |  ___/ | |  `)
	log.Info(`| |___| |_| | |_) |  __/  | |___| |__| |____) |   / ____ \| |    _| |_ `)
	log.Info(`\______\__,_|_.__/ \___|   \_____\____/|_____/   /_/    \_\_|   |_____| ©2025 Powered by Bigstack.`)
	log.Info("")
}

func printLoadedConfBody() {
	body, err := config.Opts.String()
	if err != nil {
		log.Errorf("runtime: failed to process loaded conf, the struct might be corrupted: %v", err)
		panic(err)
	}

	log.Infof("init parameters: %+v", body)
}
