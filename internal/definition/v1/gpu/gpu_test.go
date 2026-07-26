package gpu

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateGpuCardRequestUnmarshal(t *testing.T) {
	body := `{"resourceType":"sriovVgpu","profiles":[{"id":57,"count":2}]}`

	var req UpdateGpuCardRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))

	require.Equal(t, ResourceTypeSriovVgpu, req.ResourceType)
	require.Len(t, req.Profiles, 1)
	require.Equal(t, UpdateGpuCardProfile{Id: 57, Count: 2}, req.Profiles[0])
}

// hex_sdk derives a MIG-backed profile's vramMiB from `nvidia-smi mig -lgip`'s
// GiB column (GiB * 1024), so it reports a fractional MiB. Captured from cn13
// (RTX PRO 6000 Blackwell): the "1g.24gb" profile is 23.12 GiB -> 23674.88.
func TestMiBUnmarshalAcceptsFractionalValues(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want MiB
	}{
		{"23674.88", 23675}, // real cn13 value; rounds up
		{"23674.12", 23674}, // rounds down
		{"2048", 2048},      // whole values are untouched
		{"0", 0},
	} {
		var got MiB
		require.NoError(t, json.Unmarshal([]byte(tc.raw), &got), "raw %s", tc.raw)
		require.Equal(t, tc.want, got, "raw %s", tc.raw)
	}

	for _, raw := range []string{`-1`, `"2048"`, `{}`} {
		var got MiB
		require.Error(t, json.Unmarshal([]byte(raw), &got), "raw %s should be rejected", raw)
	}

	// JSON null leaves the value untouched, matching encoding/json's behaviour
	// for every other field on this struct (e.g. a null vmCountLimit).
	got := MiB(7)
	require.NoError(t, json.Unmarshal([]byte(`null`), &got))
	require.Equal(t, MiB(7), got)
}

// Regression: encoding/json aborts the whole document on one field error, so a
// single fractional migBacked vramMiB used to drop *every* profile on the card
// (the card then reported degraded with an empty profile collection).
func TestVgpuProfileCollectionSurvivesFractionalMigVram(t *testing.T) {
	raw := `{"sriov":[{"id":1518,"name":"DC-2B","vramMiB":2048,"vmCountLimit":null,"count":1,"alias":"nvidia_dc_2b_1518"}],` +
		`"migBacked":[{"id":47,"name":"1g.24gb","vramMiB":23674.88,"vmCountLimit":4,"count":0,"alias":null}]}`

	var collection VgpuProfileCollectionFromHex
	require.NoError(t, json.Unmarshal([]byte(raw), &collection))

	require.NotNil(t, collection.Sriov)
	require.Len(t, *collection.Sriov, 1)
	require.Equal(t, MiB(2048), (*collection.Sriov)[0].VramMiB)
	require.Equal(t, 1, (*collection.Sriov)[0].Count)

	require.NotNil(t, collection.MigBacked)
	require.Len(t, *collection.MigBacked, 1)
	require.Equal(t, MiB(23675), (*collection.MigBacked)[0].VramMiB)
}

func TestGpuUpdateSentinelsAreDistinct(t *testing.T) {
	for _, e := range []error{
		ErrGpuNotFound, ErrUnsupportedType, ErrProfilesNotAllowed,
		ErrProfileNotFound, ErrExceedProfileCountLimit, ErrExceedVramLimit, ErrGpuInUse,
	} {
		require.Error(t, e)
	}

	wrapped := fmt.Errorf("gpu abc is in use: %w", ErrGpuInUse)
	require.True(t, errors.Is(wrapped, ErrGpuInUse))
	require.False(t, errors.Is(wrapped, ErrGpuNotFound))
}
