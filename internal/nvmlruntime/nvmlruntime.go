// Package nvmlruntime owns the process-wide NVML lifecycle and availability
// state. It lives in its own leaf package so both the runtime (which starts and
// stops NVML) and the API handlers (which need to know whether NVML data can be
// trusted) can consult it without an import cycle.
package nvmlruntime

import (
	"errors"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "go-micro.dev/v5/logger"
)

// available is set only while nvml.Init() has succeeded; the service runs
// without NVML support otherwise, and nvml.Shutdown() must not be called in
// that case.
var available bool

// Seams for unit tests: the NVML driver is unavailable there.
var (
	nvmlInit     = nvml.Init
	nvmlShutdown = nvml.Shutdown
)

// Init brings up NVML. NVML is optional (nodes without NVIDIA GPUs run the same
// service), so a missing driver degrades to running without NVML support rather
// than aborting. Other failures also degrade so the service still boots, but
// they are logged as real errors and leave availability false so handlers can
// tell that GPU runtime data is untrustworthy.
func Init() {
	switch ret := nvmlInit(); ret {
	case nvml.SUCCESS:
		available = true
	case nvml.ERROR_DRIVER_NOT_LOADED, nvml.ERROR_LIBRARY_NOT_FOUND:
		log.Warnf("NVIDIA driver is not loaded. Continuing without NVML support.")
	default:
		log.Errorf("Failed to initialize NVML: %v. Continuing without NVML support; GPU runtime data will be reported as degraded.", nvml.ErrorString(ret))
	}
}

// Shutdown tears down NVML if it was initialized.
func Shutdown() error {
	if !available {
		return nil
	}

	ret := nvmlShutdown()
	if ret != nvml.SUCCESS {
		errorString := nvml.ErrorString(ret)
		log.Errorf("Failed to shutdown NVML: %v", errorString)
		return errors.New(errorString)
	}

	available = false
	log.Infof("NVML is shut down")

	return nil
}

// IsAvailable reports whether NVML initialized successfully. When false, any
// GPU that hex reports as NVML-managed cannot be enriched with runtime stats or
// attachment data, so its capacity numbers must be treated as degraded.
func IsAvailable() bool {
	return available
}
