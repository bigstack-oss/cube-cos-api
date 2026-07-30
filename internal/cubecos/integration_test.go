package cubecos

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/stretchr/testify/require"
)

// hex_sdk emits OpenStack UTC — 15:02:37 in +08:00, 02:02:37 in -05:00.
const sourceUpdateTime = "2026-07-24T07:02:37Z"

// Stands in for the value runtime.initIdentity reads once at startup. It is
// already in the node's local offset, so tests only need it to be recognisable.
const firmwareUpdatedAt = "2026-07-01T09:00:00+08:00"

// Pins the startup-derived firmware timestamp and restores it so the value does
// not leak between tests.
func stubActiveFirmwareUpdatedAt(t *testing.T, updatedAt string) {
	t.Helper()

	orig := base.ActiveFirmwareUpdatedAt
	base.ActiveFirmwareUpdatedAt = updatedAt

	t.Cleanup(func() { base.ActiveFirmwareUpdatedAt = orig })
}

func TestConvertStorageTimesReportsTheSourceTimeInTheNodeLocalOffset(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
		want          string
	}{
		{name: "east of UTC", offsetSeconds: 8 * 60 * 60, want: "2026-07-24T15:02:37+08:00"},
		{name: "west of UTC", offsetSeconds: -5 * 60 * 60, want: "2026-07-24T02:02:37-05:00"},
		{name: "at UTC", offsetSeconds: 0, want: "2026-07-24T07:02:37Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useLocalZone(t, test.offsetSeconds)
			stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

			list := []storages.Cinder{{Name: "vsan", UpdateTime: sourceUpdateTime}}
			convertStorageTimes(&list)

			require.Equal(t, test.want, list[0].UpdateTime)
		})
	}
}

// A storage hex_sdk reports no update time for has no timestamp of its own, so
// it reports the same stable placeholder a built-in storage does. Substituting
// request time would move the value on every GET.
func TestConvertStorageTimesReportsAStableTimeWhenTheSourceHasNone(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
	}{
		{name: "east of UTC", offsetSeconds: 8 * 60 * 60},
		{name: "west of UTC", offsetSeconds: -5 * 60 * 60},
		{name: "at UTC", offsetSeconds: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useLocalZone(t, test.offsetSeconds)
			stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

			first := []storages.Cinder{{Name: "vsan"}}
			second := []storages.Cinder{{Name: "vsan"}}
			convertStorageTimes(&first)
			convertStorageTimes(&second)

			require.Equal(t, firmwareUpdatedAt, first[0].UpdateTime)
			require.Equal(t, first[0].UpdateTime, second[0].UpdateTime)
		})
	}
}

func TestConvertStorageTimesKeepsBuiltInStoragesOnTheFirmwareTime(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

	list := []storages.Cinder{{Name: "CubeStorage", IsBuiltIn: true, UpdateTime: sourceUpdateTime}}
	convertStorageTimes(&list)

	require.Equal(t, firmwareUpdatedAt, list[0].UpdateTime)
}

func TestConvertStorageTimesLeavesAnUnparseableSourceTimeAlone(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

	list := []storages.Cinder{{Name: "vsan", UpdateTime: "not a timestamp"}}
	convertStorageTimes(&list)

	require.Equal(t, "not a timestamp", list[0].UpdateTime)
}

// The details response emits the top-level update time and the one nested under
// "storage". Both arrive raw UTC, so both need converting.
func TestConvertStorageDetailsTimesConvertsBothUpdateTimeFields(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
		want          string
	}{
		{name: "east of UTC", offsetSeconds: 8 * 60 * 60, want: "2026-07-24T15:02:37+08:00"},
		{name: "west of UTC", offsetSeconds: -5 * 60 * 60, want: "2026-07-24T02:02:37-05:00"},
		{name: "at UTC", offsetSeconds: 0, want: "2026-07-24T07:02:37Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useLocalZone(t, test.offsetSeconds)
			stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

			details := &storages.CinderDetails{
				Name:       "vsan",
				UpdateTime: sourceUpdateTime,
				Storage:    storages.Storage{UpdateTime: sourceUpdateTime},
			}
			convertStorageDetailsTimes(details)

			require.Equal(t, test.want, details.UpdateTime)
			require.Equal(t, test.want, details.Storage.UpdateTime)
		})
	}
}

// The list and details endpoints report the same storage, so they must agree on
// the instant and the offset.
func TestConvertStorageDetailsTimesAgreesWithTheListPath(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
	}{
		{name: "east of UTC", offsetSeconds: 8 * 60 * 60},
		{name: "west of UTC", offsetSeconds: -5 * 60 * 60},
		{name: "at UTC", offsetSeconds: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useLocalZone(t, test.offsetSeconds)
			stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

			list := []storages.Cinder{{Name: "vsan", UpdateTime: sourceUpdateTime}}
			details := &storages.CinderDetails{
				Name:       "vsan",
				UpdateTime: sourceUpdateTime,
				Storage:    storages.Storage{UpdateTime: sourceUpdateTime},
			}

			convertStorageTimes(&list)
			convertStorageDetailsTimes(details)

			require.Equal(t, list[0].UpdateTime, details.UpdateTime)
			require.Equal(t, list[0].UpdateTime, details.Storage.UpdateTime)
		})
	}
}

// updateTime is required in GetIntegrationStorageResponse, so the details path
// needs the same stable placeholder rather than an empty string.
func TestConvertStorageDetailsTimesReportsAStableTimeWhenTheSourceHasNone(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

	details := &storages.CinderDetails{Name: "vsan"}
	convertStorageDetailsTimes(details)

	require.Equal(t, firmwareUpdatedAt, details.UpdateTime)
	require.Equal(t, firmwareUpdatedAt, details.Storage.UpdateTime)
}

func TestConvertStorageDetailsTimesKeepsBuiltInStoragesOnTheFirmwareTime(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubActiveFirmwareUpdatedAt(t, firmwareUpdatedAt)

	details := &storages.CinderDetails{
		Name:       "CubeStorage",
		IsBuiltIn:  true,
		UpdateTime: sourceUpdateTime,
		Storage:    storages.Storage{UpdateTime: sourceUpdateTime},
	}
	convertStorageDetailsTimes(details)

	require.Equal(t, firmwareUpdatedAt, details.UpdateTime)
	require.Equal(t, firmwareUpdatedAt, details.Storage.UpdateTime)
}

func TestConvertStorageDetailsTimesToleratesANilStorage(t *testing.T) {
	useLocalZone(t, 8*60*60)

	require.NotPanics(t, func() { convertStorageDetailsTimes(nil) })
}
