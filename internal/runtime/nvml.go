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

		if ret == nvml.ERROR_DRIVER_NOT_LOADED {
			log.Warnf("Failed to initialize NVML: %v", errorString)
			log.Warnf("NVIDIA driver is not loaded. Continuing without NVML support, GPU features will be unavailable.")

			return nil
		}

		log.Fatalf("Failed to initialize NVML: %v", errorString)

		return errors.New(errorString)
	}

	return nil
}
