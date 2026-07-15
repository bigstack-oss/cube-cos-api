package runtime

import (
	"github.com/bigstack-oss/cube-cos-api/internal/nvmlruntime"
)

// initNvml and ShutdownNvml are thin wrappers over the nvmlruntime leaf package,
// which owns the NVML lifecycle and availability state shared with the API
// handlers.
func initNvml() {
	nvmlruntime.Init()
}

func ShutdownNvml() error {
	return nvmlruntime.Shutdown()
}
