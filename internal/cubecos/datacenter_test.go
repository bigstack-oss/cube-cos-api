package cubecos

import (
	"errors"
	"testing"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/stretchr/testify/require"
)

// Pins the process-wide local zone the way runtime.initSystemTime would, and
// restores it so the offset does not leak between tests.
func useLocalZone(t *testing.T, offsetSeconds int) {
	t.Helper()

	origSeconds := time.LocalZoneSeconds
	origZone := time.LocalFixedZone

	time.LocalZoneSeconds = offsetSeconds
	time.LocalFixedZone = ostime.FixedZone("", offsetSeconds)

	t.Cleanup(func() {
		time.LocalZoneSeconds = origSeconds
		time.LocalFixedZone = origZone
	})
}

// Restores the fixpack-history seam so stubs do not leak between tests.
func stubLastInstalledFixpack(t *testing.T, fn func() ([]string, error)) {
	t.Helper()

	orig := getLastInstalledFixpack
	getLastInstalledFixpack = fn

	t.Cleanup(func() { getLastInstalledFixpack = orig })
}

// hex_config fixpack_get_history emits a bare wall clock with no zone. It is
// the node's local time, so the emitted value keeps those digits and gains the
// node's offset — it must not be re-shifted as if it had been UTC.
func TestGetFixpackUpdatedAtKeepsTheWallClockAndAddsTheLocalOffset(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
		want          string
	}{
		{name: "east of UTC", offsetSeconds: 8 * 60 * 60, want: "2025-07-24T15:02:37+08:00"},
		{name: "west of UTC", offsetSeconds: -5 * 60 * 60, want: "2025-07-24T15:02:37-05:00"},
		{name: "at UTC", offsetSeconds: 0, want: "2025-07-24T15:02:37Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useLocalZone(t, test.offsetSeconds)
			stubLastInstalledFixpack(t, func() ([]string, error) {
				return []string{"24 Jul 2025 15:02:37", "3.2.0", "fixpack-name", "", "", "", ""}, nil
			})

			got, err := GetFixpackUpdatedAt()

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

// The fixpack list endpoint formats the same hex_config value via
// convertRawTime. Both endpoints report the same fixpack, so they must agree.
func TestGetFixpackUpdatedAtAgreesWithTheFixpackListPath(t *testing.T) {
	useLocalZone(t, 8*60*60)

	const rawUpdatedAt = "24 Jul 2025 15:02:37"
	stubLastInstalledFixpack(t, func() ([]string, error) {
		return []string{rawUpdatedAt, "3.2.0", "fixpack-name", "", "", "", ""}, nil
	})

	got, err := GetFixpackUpdatedAt()

	require.NoError(t, err)
	require.Equal(t, convertRawTime(time.FormatFixpack, rawUpdatedAt), got)
}

func TestGetFixpackUpdatedAtRejectsAnUnparseableWallClock(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubLastInstalledFixpack(t, func() ([]string, error) {
		return []string{"not a timestamp", "3.2.0", "fixpack-name", "", "", "", ""}, nil
	})

	_, err := GetFixpackUpdatedAt()

	require.ErrorContains(t, err, "failed to parse fixpack updatedAt")
}

func TestGetFixpackUpdatedAtPropagatesTheHistoryLookupError(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubLastInstalledFixpack(t, func() ([]string, error) {
		return nil, errors.New("hex_config unavailable")
	})

	_, err := GetFixpackUpdatedAt()

	require.ErrorContains(t, err, "hex_config unavailable")
}
