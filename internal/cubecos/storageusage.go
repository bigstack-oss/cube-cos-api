package cubecos

import (
	"strconv"
	"strings"
)

// Bounds and default of storage.usage.interval (cubecos config_cinder.cpp).
const (
	StorageUsageInterval = "storage.usage.interval"

	storageUsageIntervalDefault = 15
	storageUsageIntervalMin     = 5
	storageUsageIntervalMax     = 1440

	storageUsageHistoryMinMinutes = 60
)

var readStorageUsageInterval = func() (string, error) {
	return GetTuningValue(StorageUsageInterval)
}

// StorageUsageIntervalMinutes returns the storage usage sampling interval,
// falling back to the default when the tuning cannot be read.
func StorageUsageIntervalMinutes() int {
	raw, err := readStorageUsageInterval()
	if err != nil {
		return storageUsageIntervalDefault
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return storageUsageIntervalDefault
	}

	switch {
	case minutes < storageUsageIntervalMin:
		return storageUsageIntervalMin
	case minutes > storageUsageIntervalMax:
		return storageUsageIntervalMax
	}
	return minutes
}

// StorageUsageRankMinutes is the rank lookback: two sampling intervals, so the
// latest sample is always in range and a stopped guest agent drops out soon.
func StorageUsageRankMinutes() int {
	return 2 * StorageUsageIntervalMinutes()
}

// StorageUsageHistoryMinutes is the history lookback: four sampling intervals,
// and never less than the hour the other VM metrics show.
func StorageUsageHistoryMinutes() int {
	minutes := 4 * StorageUsageIntervalMinutes()
	if minutes < storageUsageHistoryMinMinutes {
		return storageUsageHistoryMinMinutes
	}
	return minutes
}
