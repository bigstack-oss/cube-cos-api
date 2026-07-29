package nodes

import (
	"testing"
	ostime "time"

	bstime "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/stretchr/testify/require"
)

// Pins the process-wide local zone the way runtime.initSystemTime would, and
// restores it so the offset does not leak between tests.
func useLocalZone(t *testing.T, offsetSeconds int) {
	t.Helper()

	origSeconds := bstime.LocalZoneSeconds
	origZone := bstime.LocalFixedZone

	bstime.LocalZoneSeconds = offsetSeconds
	bstime.LocalFixedZone = ostime.FixedZone("", offsetSeconds)

	t.Cleanup(func() {
		bstime.LocalZoneSeconds = origSeconds
		bstime.LocalFixedZone = origZone
	})
}

// The BMC reports a bare wall clock with no zone. It is the node's local time,
// so the emitted value must carry the node's offset and keep the same digits.
func TestConvertBmcTimeKeepsTheWallClockAndAddsTheLocalOffset(t *testing.T) {
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

			got, err := convertBmcTime("Thu Jul 24 15:02:37 2025")

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestConvertBmcTimeRejectsAnUnparseableWallClock(t *testing.T) {
	useLocalZone(t, 8*60*60)

	_, err := convertBmcTime("Unspecified")

	require.Error(t, err)
}
