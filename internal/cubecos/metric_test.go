package cubecos

import (
	"encoding/json"
	osmath "math"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/stretchr/testify/require"
)

// A metric gap gives an empty stat slice. The average must stay 0, because
// json.Marshal rejects NaN and answers with no body at all.
func TestHostsCpuAverageWithoutStatsStaysZero(t *testing.T) {
	average := GetHostsCpuAverage([]metric.Compute{})

	require.False(t, osmath.IsNaN(average.UsedPercent), "usedPercent is NaN")
	require.False(t, osmath.IsNaN(average.FreePercent), "freePercent is NaN")
	require.Equal(t, float64(0), average.UsedPercent)
	require.Equal(t, float64(0), average.FreePercent)

	_, err := json.Marshal(average)
	require.NoError(t, err)
}

func TestHostsCpuAverageWithStatsKeepsAveraging(t *testing.T) {
	average := GetHostsCpuAverage([]metric.Compute{
		{TotalCores: 8, UsedCores: 2, UsedPercent: 25, FreeCores: 6, FreePercent: 75},
		{TotalCores: 8, UsedCores: 6, UsedPercent: 75, FreeCores: 2, FreePercent: 25},
	})

	require.Equal(t, float64(16), average.TotalCores)
	require.Equal(t, float64(8), average.UsedCores)
	require.Equal(t, float64(50), average.UsedPercent)
	require.Equal(t, float64(50), average.FreePercent)
}

func TestHostsMemoryAverageWithoutStatsStaysZero(t *testing.T) {
	average := GetHostsMemoryAverage([]metric.Space{})

	require.False(t, osmath.IsNaN(average.UsedPercent), "usedPercent is NaN")
	require.False(t, osmath.IsNaN(average.FreePercent), "freePercent is NaN")
	require.Equal(t, float64(0), average.UsedPercent)
	require.Equal(t, float64(0), average.FreePercent)

	_, err := json.Marshal(average)
	require.NoError(t, err)
}

// InfluxDB holds no disk row inside the query window, so the sizes stay at 0.
// Both percent fields must stay 0 as well.
func TestSpaceUsagePercentWithoutTotalStaysZero(t *testing.T) {
	space := &metric.Space{}

	setSpaceUsagePercent(space)

	require.False(t, osmath.IsNaN(space.UsedPercent), "usedPercent is NaN")
	require.False(t, osmath.IsNaN(space.FreePercent), "freePercent is NaN")
	require.Equal(t, float64(0), space.UsedPercent)
	require.Equal(t, float64(0), space.FreePercent)

	_, err := json.Marshal(space)
	require.NoError(t, err)
}

func TestSpaceUsagePercentWithTotalFillsBothPercents(t *testing.T) {
	space := &metric.Space{TotalMiB: 1000, UsedMiB: 250, FreeMiB: 750}

	setSpaceUsagePercent(space)

	require.Equal(t, float64(25), space.UsedPercent)
	require.Equal(t, float64(75), space.FreePercent)
}

func TestHostsMemoryAverageWithStatsKeepsAveraging(t *testing.T) {
	average := GetHostsMemoryAverage([]metric.Space{
		{TotalMiB: 1024, UsedMiB: 256, UsedPercent: 25, FreeMiB: 768, FreePercent: 75},
		{TotalMiB: 1024, UsedMiB: 768, UsedPercent: 75, FreeMiB: 256, FreePercent: 25},
	})

	require.Equal(t, float64(2048), average.TotalMiB)
	require.Equal(t, float64(1024), average.UsedMiB)
	require.Equal(t, float64(50), average.UsedPercent)
	require.Equal(t, float64(50), average.FreePercent)
}
