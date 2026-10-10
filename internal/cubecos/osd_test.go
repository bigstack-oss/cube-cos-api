package cubecos

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/stretchr/testify/require"
)

// The UI's OSD Reweight field is the 0..1 override weight, so it must go to
// `ceph osd reweight`. `ceph osd crush reweight` would overwrite the CRUSH
// weight (the disk size in TiB) and move data (#1772).
func TestReweightOsdArgsUsesOsdReweightNotCrushReweight(t *testing.T) {
	args := reweightOsdArgs(nodes.OsdReqOpts{OsdId: "osd.2", Reweight: 0.95})
	require.Equal(t, []string{"osd", "reweight", "osd.2", "0.950000"}, args)
	require.NotContains(t, args, "crush")
}
