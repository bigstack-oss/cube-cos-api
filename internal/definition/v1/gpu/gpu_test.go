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
