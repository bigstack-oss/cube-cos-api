package nvmlruntime

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	origNvmlInit := nvmlInit
	t.Cleanup(func() {
		nvmlInit = origNvmlInit
		available = false
	})

	t.Run("success marks nvml as available", func(t *testing.T) {
		available = false
		nvmlInit = func() nvml.Return { return nvml.SUCCESS }

		Init()

		require.True(t, IsAvailable())
	})

	t.Run("missing driver continues without nvml", func(t *testing.T) {
		available = false
		nvmlInit = func() nvml.Return { return nvml.ERROR_DRIVER_NOT_LOADED }

		Init()

		require.False(t, IsAvailable())
	})

	t.Run("unexpected failure continues without nvml", func(t *testing.T) {
		available = false
		nvmlInit = func() nvml.Return { return nvml.ERROR_UNKNOWN }

		Init()

		require.False(t, IsAvailable())
	})
}

func TestShutdown(t *testing.T) {
	origShutdown := nvmlShutdown
	t.Cleanup(func() {
		nvmlShutdown = origShutdown
		available = false
	})

	t.Run("no-op when nvml was never available", func(t *testing.T) {
		available = false
		called := false
		nvmlShutdown = func() nvml.Return { called = true; return nvml.SUCCESS }

		require.NoError(t, Shutdown())
		require.False(t, called)
	})

	t.Run("shuts down and clears availability", func(t *testing.T) {
		available = true
		nvmlShutdown = func() nvml.Return { return nvml.SUCCESS }

		require.NoError(t, Shutdown())
		require.False(t, IsAvailable())
	})
}
