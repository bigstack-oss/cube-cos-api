package time

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 2025-07-24T07:02:37Z — 15:02:37 in +08:00, 02:02:37 in -05:00.
const testBootEpoch = uint64(1753340557)

// Pins the process-wide local zone the way runtime.initSystemTime would, and
// restores it so the offset does not leak between tests.
func useLocalZone(t *testing.T, offsetSeconds int) {
	t.Helper()

	origSeconds := LocalZoneSeconds
	origZone := LocalFixedZone

	LocalZoneSeconds = offsetSeconds
	LocalFixedZone = time.FixedZone("", offsetSeconds)

	t.Cleanup(func() {
		LocalZoneSeconds = origSeconds
		LocalFixedZone = origZone
	})
}

// Restores the bootTime seam so stubs do not leak between tests.
func stubBootTime(t *testing.T, fn func() (uint64, error)) {
	t.Helper()

	orig := bootTime
	bootTime = fn

	t.Cleanup(func() { bootTime = orig })
}

func TestBootReportsTheBootInstantInTheNodeLocalOffset(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
		want          string
	}{
		{name: "east of UTC", offsetSeconds: 8 * 60 * 60, want: "2025-07-24T15:02:37+08:00"},
		{name: "west of UTC", offsetSeconds: -5 * 60 * 60, want: "2025-07-24T02:02:37-05:00"},
		{name: "at UTC", offsetSeconds: 0, want: "2025-07-24T07:02:37Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useLocalZone(t, test.offsetSeconds)
			stubBootTime(t, func() (uint64, error) { return testBootEpoch, nil })

			require.Equal(t, test.want, Boot())
		})
	}
}

func TestBootFallsBackToNowInTheNodeLocalOffset(t *testing.T) {
	useLocalZone(t, 8*60*60)
	stubBootTime(t, func() (uint64, error) { return 0, errors.New("boot time unavailable") })

	parsed, err := time.Parse(FormatRFC3339, Boot())
	require.NoError(t, err)

	_, offset := parsed.Zone()
	require.Equal(t, 8*60*60, offset)
	require.WithinDuration(t, time.Now(), parsed, time.Minute)
}
