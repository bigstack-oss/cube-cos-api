package runtime

import (
	"errors"

	log "go-micro.dev/v5/logger"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// Set only when nvml.Init() succeeds; the service starts without NVML
// support otherwise, and nvml.Shutdown() must not be called in that case.
var isNvmlInitialized bool

// Seam for unit tests: the NVML driver is unavailable there.
var nvmlInit = nvml.Init

// NVML is optional (nodes without NVIDIA GPUs run the same service), so any
// init failure degrades to running without NVML support instead of aborting.
func initNvml() {
	switch ret := nvmlInit(); ret {
	case nvml.SUCCESS:
		isNvmlInitialized = true
	case nvml.ERROR_DRIVER_NOT_LOADED, nvml.ERROR_LIBRARY_NOT_FOUND:
		log.Warnf("NVIDIA driver is not loaded. Continuing without NVML support.")
	default:
		log.Warnf("Failed to initialize NVML: %v. Continuing without NVML support.", nvml.ErrorString(ret))
	}
}

func ShutdownNvml() error {
	if !isNvmlInitialized {
		return nil
	}

	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		errorString := nvml.ErrorString(ret)
		log.Errorf("Failed to shutdown NVML: %v", errorString)
		return errors.New(errorString)
	}

	isNvmlInitialized = false
	log.Infof("NVML is shut down")

	return nil
}
