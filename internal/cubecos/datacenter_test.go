package cubecos

import (
	"errors"
	"testing"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
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

// Restores the fixpack-history seam so stubs do not leak between tests.
func stubFixpackHistory(t *testing.T, fn func() ([]fixpacks.Fixpack, error)) {
	t.Helper()

	orig := listFixpackHistory
	listFixpackHistory = fn

	t.Cleanup(func() { listFixpackHistory = orig })
}

// Feeds raw hex_config fixpack_get_history output through the real parser.
func stubRawFixpackHistory(t *testing.T, raw string) {
	t.Helper()

	stubFixpackHistory(t, func() ([]fixpacks.Fixpack, error) {
		return convertHistoryToFixpacks([]byte(raw))
	})
}

func TestGetCurrentFixpackReportsTheLatestInstalledFixpack(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubRawFixpackHistory(t, ""+
		"01 May 2026 09:00:00|001|first-fix|Yes|installed|note|\n"+
		"06 May 2026 14:24:29|002|appctl 4eaec97|No|installed|fix appfw offline install|\n")

	got, err := GetCurrentFixpack()

	require.NoError(t, err)
	require.Equal(t, base.Fixpack{
		Name:      "appctl 4eaec97",
		Version:   "002",
		UpdatedAt: "2026-05-06T14:24:29+08:00",
	}, got)
}

// A rollback appends an "uninstalled" line rather than removing the install,
// so the last history line is not necessarily the fixpack in effect.
func TestGetCurrentFixpackFallsBackToThePreviousFixpackAfterARollback(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubRawFixpackHistory(t, ""+
		"01 May 2026 09:00:00|001|first-fix|Yes|installed|note|\n"+
		"06 May 2026 14:24:29|002|second-fix|Yes|installed|note|\n"+
		"07 May 2026 10:00:00|002|second-fix|Yes|uninstalled|note|\n")

	got, err := GetCurrentFixpack()

	require.NoError(t, err)
	require.Equal(t, "001", got.Version)
	require.Equal(t, "first-fix", got.Name)
}

func TestGetCurrentFixpackIsEmptyWhenEveryFixpackIsRolledBack(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubRawFixpackHistory(t, ""+
		"06 May 2026 14:24:29|001|first-fix|Yes|installed|note|\n"+
		"07 May 2026 10:00:00|001|first-fix|Yes|uninstalled|note|\n")

	got, err := GetCurrentFixpack()

	require.NoError(t, err)
	require.Equal(t, base.Fixpack{}, got)
}

func TestGetCurrentFixpackIsEmptyWithoutHistory(t *testing.T) {
	stubRawFixpackHistory(t, "")

	got, err := GetCurrentFixpack()

	require.NoError(t, err)
	require.Equal(t, base.Fixpack{}, got)
}

func TestGetCurrentFixpackPropagatesTheHistoryLookupError(t *testing.T) {
	stubFixpackHistory(t, func() ([]fixpacks.Fixpack, error) {
		return nil, errors.New("hex_config unavailable")
	})

	_, err := GetCurrentFixpack()

	require.ErrorContains(t, err, "hex_config unavailable")
}
