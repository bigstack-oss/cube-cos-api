package cubecos

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// hexConfigRefusalFixture is real `hex_config -e gpu_resource_set` output
// captured on cn13 (RTX PRO 6000 Blackwell, 2026-09-09) when the profiles
// argument was not a JSON array. Kept verbatim, including the "Executing:"
// trace line, because separating that from the two Error lines is the whole
// job of HexErrorReason.
const hexConfigRefusalFixture = `hex_config[501827]: Executing: /usr/sbin/hex_sdk gpu_resource_set_check GPU-0a5ba9ad-a575-72c7-d51d-3e726232cfd1 sriovVgpu not-json
hex_config[501827]: Error: gpu_resource_set: profiles must be a non-empty JSON array of { id, count } objects
hex_config[501827]: Error: gpu_resource_set: pre-condition check failed for GPU GPU-0a5ba9ad-a575-72c7-d51d-3e726232cfd1
`

func TestHexErrorReasonKeepsOnlyTheErrorLines(t *testing.T) {
	reason := HexErrorReason([]byte(hexConfigRefusalFixture), errors.New("exit status 1"))

	require.Equal(t,
		"gpu_resource_set: profiles must be a non-empty JSON array of { id, count } objects; "+
			"gpu_resource_set: pre-condition check failed for GPU GPU-0a5ba9ad-a575-72c7-d51d-3e726232cfd1",
		reason)
	require.NotContains(t, reason, "Executing:")
}

// Without -e, hex_config exits 1 having written nothing at all. That is the
// state #1452 was reported in, and the reason the fallback chain has to end
// somewhere useful rather than returning an empty string.
func TestHexErrorReasonFallsBackToTheExecError(t *testing.T) {
	require.Equal(t, "exit status 1", HexErrorReason(nil, errors.New("exit status 1")))
	require.Equal(t, "exit status 1", HexErrorReason([]byte("  \n \n"), errors.New("exit status 1")))
}

func TestHexErrorReasonFallsBackToRawOutput(t *testing.T) {
	require.Equal(t, "something went wrong",
		HexErrorReason([]byte("something went wrong\n"), errors.New("exit status 1")))
}

func TestHexErrorReasonNeverReturnsEmpty(t *testing.T) {
	require.Equal(t, "no reason reported", HexErrorReason(nil, nil))
}
