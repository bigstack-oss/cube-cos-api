package runtime

import (
	log "go-micro.dev/v5/logger"
)

func showPromptMessage() {
	log.Info("")
	log.Info(`  _____      _                      _____ _____`)
	log.Info(` / ____|    | |               /\   |  __ \_   _|`)
	log.Info(`| |    _   _| |__   ___      /  \  | |__) || |  `)
	log.Info(`| |   | | | | '_ \ / _ \    / /\ \ |  ___/ | |  `)
	log.Info(`| |___| |_| | |_) |  __/   / ____ \| |    _| |_ `)
	log.Info(` \_____\__,_|_.__/ \___|  /_/    \_\_|   |_____| | ©2024 Powered by Bigstack.`)
	log.Info("")
}
