package main

import (
	_ "github.com/bigstack-oss/cube-cos-api/api"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/runtime"
	svc "github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func main() {
	err := config.SyncOptions()
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return
	}

	srv, err := runtime.NewHttpServer()
	if err != nil {
		log.Errorf("failed to init runtime: %v", err)
		return
	}

	err = svc.Micro(srv).Run()
	if err != nil {
		log.Errorf("failed to run service: %v", err)
	}
}
