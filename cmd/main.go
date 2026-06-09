package main

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	_ "github.com/bigstack-oss/cube-cos-api/api"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/runtime"
	svc "github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func nvmlTest() {
	// 1. 初始化 NVML (這是最重要的起手式)
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("無法初始化 NVML: %v", nvml.ErrorString(ret))
	}
	// 記得在程式結束時關閉它
	defer nvml.Shutdown()
}

func main() {
	err := config.SyncOptions()

	if err != nil {
		log.Errorf("failed to load config(%v)", err)
		return
	}

	srv, err := runtime.NewHttpServer()
	if err != nil {
		log.Errorf("failed to init runtime(%v)", err)
		return
	}

	err = svc.Micro(srv).Run()
	if err != nil {
		log.Errorf("failed to run service(%v)", err)
	}
}
