package ceph

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// osdDfFixture is the first node of `ceph osd df osd.2 -f json` on a lab node
// (3.6 TB HDD split into two OSDs), reduced to the fields that matter here.
// Its CRUSH weight is above 1 while its reweight is 1, which is what #1772
// tripped over.
const osdDfFixture = `{"nodes":[{"name":"osd.2","crush_weight":1.8189849853515625,"reweight":1,"device_class":"hdd"}]}`

func TestConvertToOsdReportsReweightNotCrushWeight(t *testing.T) {
	raw := RawOsd{}
	require.NoError(t, json.Unmarshal([]byte(osdDfFixture), &raw))

	osd := convertToOsd(raw.Nodes[0])

	require.Equal(t, "osd.2", osd.Id)
	require.Equal(t, "HDD", osd.DeviceClass)
	require.Equal(t, 1.0, osd.Reweight)
}

// Ceph stores the reweight as 16.16 fixed point and truncates on write, so
// `ceph osd reweight osd.N 0.95` reads back from `osd df` as 0.94999695.
func TestConvertToOsdRoundsReweightToTheNearestHundredth(t *testing.T) {
	osd := convertToOsd(Node{Name: "osd.3", CrushWeight: 1.81898, Reweight: 0.9499969482421875})
	require.Equal(t, 0.95, osd.Reweight)
}

// Every value the UI accepts (0.00..1.00) must read back as itself after a
// round trip through Ceph's fixed point.
func TestEveryTwoDecimalReweightSurvivesCephFixedPoint(t *testing.T) {
	for k := 0; k <= 100; k++ {
		want := float64(k) / 100
		stored := float64(int(want*0x10000)) / 0x10000

		osd := convertToOsd(Node{Reweight: stored})

		require.Equal(t, want, osd.Reweight, "reweight %.2f stored as %v", want, stored)
	}
}

func TestGetMaxOsdReweight(t *testing.T) {
	tests := []struct {
		name string
		osds []Osd
		want float64
	}{
		{"no osds", nil, 0.0},
		{"one osd", []Osd{{Reweight: 0.5}}, 0.5},
		{"max of several", []Osd{{Reweight: 0.5}, {Reweight: 0.95}, {Reweight: 0.7}}, 0.95},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getMaxOsdReweight(tt.osds))
		})
	}
}
