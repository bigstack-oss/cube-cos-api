package nodes

import (
	"testing"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ipmi"
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

// The BMC is accepted only when its FRU board serial is the node's board
// serial, never when either side is empty or one merely contains the other.
func TestMatchBoardSerialAcceptsOnlyTheNodesOwnBoard(t *testing.T) {
	tests := []struct {
		name          string
		nodeSerial    string
		boardSerial   string
		productSerial string
		wantErr       string
	}{
		{name: "exact plain serial", nodeSerial: "ABCD123456", boardSerial: "ABCD123456"},
		{name: "exact with surrounding whitespace", nodeSerial: " ABCD123456\n", boardSerial: "ABCD123456 "},
		{name: "dotted product and board serial", nodeSerial: ".SVCTAG1.CN1234567890AB.", boardSerial: "CN1234567890AB", productSerial: "SVCTAG1"},
		{name: "empty BMC serial", nodeSerial: "ABCD123456", boardSerial: "", wantErr: "reports no board serial"},
		{name: "whitespace BMC serial", nodeSerial: "ABCD123456", boardSerial: "   ", wantErr: "reports no board serial"},
		{name: "empty node serial", nodeSerial: "", boardSerial: "ABCD123456", wantErr: "has no board serial on record"},
		{name: "both empty", nodeSerial: "", boardSerial: "", wantErr: "board serial is mismatched"},
		{name: "prefix of node serial", nodeSerial: "ABCD123456", boardSerial: "ABCD12", wantErr: "board serial is mismatched"},
		{name: "suffix of node serial", nodeSerial: "ABCD123456", boardSerial: "123456", wantErr: "board serial is mismatched"},
		{name: "infix of node serial", nodeSerial: "ABCD123456", boardSerial: "D1234", wantErr: "board serial is mismatched"},
		{name: "dotted with a partial board component", nodeSerial: ".SVCTAG1.CN1234567890AB.", boardSerial: "CN12345", productSerial: "SVCTAG1", wantErr: "board serial is mismatched"},
		{name: "dotted with no product serial", nodeSerial: ".SVCTAG1.CN1234567890AB.", boardSerial: "CN1234567890AB", wantErr: "board serial is mismatched"},
		{name: "dotted with another product serial", nodeSerial: ".SVCTAG1.CN1234567890AB.", boardSerial: "CN1234567890AB", productSerial: "SVCTAG2", wantErr: "board serial is mismatched"},
		{name: "unrelated serial", nodeSerial: "ABCD123456", boardSerial: "RT694ZZ999", wantErr: "board serial is mismatched"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fru := &ipmi.FRU{}
			fru.Board.Serial = test.boardSerial
			fru.Product.Serial = test.productSerial

			err := matchBoardSerial("node1", test.nodeSerial, "192.0.2.10", fru)

			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.Contains(t, err.Error(), test.wantErr)
		})
	}
}
