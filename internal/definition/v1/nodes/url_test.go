package nodes

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/stretchr/testify/require"
)

func TestUpdateGpuCardUrl(t *testing.T) {
	base.DataCenterName = "dc1"

	n := &Node{Protocol: "https", Address: "10.0.0.5:8443", Hostname: "compute-1"}

	require.Equal(t,
		"https://10.0.0.5:8443/api/v1/datacenters/dc1/nodes/compute-1/gpuCards/GPU-abc",
		n.UpdateGpuCardUrl("GPU-abc"),
	)
}
