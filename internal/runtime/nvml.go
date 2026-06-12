package runtime

import (
	"errors"

	log "go-micro.dev/v5/logger"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

func initNvml() error {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		errorString := nvml.ErrorString(ret)
		log.Fatalf("Failed to initialize NVML: %v", errorString)
		return errors.New(errorString)
	}
	defer nvml.Shutdown()
	return nil
}
