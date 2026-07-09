package runtime

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/stretchr/testify/require"
)

func TestInitNvml(t *testing.T) {
	origNvmlInit := nvmlInit
	t.Cleanup(func() {
		nvmlInit = origNvmlInit
		isNvmlInitialized = false
	})

	t.Run("success marks nvml as initialized", func(t *testing.T) {
		isNvmlInitialized = false
		nvmlInit = func() nvml.Return { return nvml.SUCCESS }

		initNvml()

		require.True(t, isNvmlInitialized)
	})

	t.Run("missing driver continues without nvml", func(t *testing.T) {
		isNvmlInitialized = false
		nvmlInit = func() nvml.Return { return nvml.ERROR_DRIVER_NOT_LOADED }

		initNvml()

		require.False(t, isNvmlInitialized)
	})

	t.Run("unexpected failure continues without nvml", func(t *testing.T) {
		isNvmlInitialized = false
		nvmlInit = func() nvml.Return { return nvml.ERROR_UNKNOWN }

		initNvml()

		require.False(t, isNvmlInitialized)
	})
}
