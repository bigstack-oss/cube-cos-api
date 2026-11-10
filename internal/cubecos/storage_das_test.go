package cubecos

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

// `lsblk --json -o NAME,TYPE,TRAN -e 43` on a node with Dell Compellent LUNs over
// two FC paths each: sdg/sdae and sdr/sdap are the two paths of one LUN
const lsblkTreeWithFcMultipath = `{
   "blockdevices": [
      {"name": "sda", "type": "disk", "tran": null, "children": [
         {"name": "sda1", "type": "part", "tran": null},
         {"name": "sda5", "type": "part", "tran": null}
      ]},
      {"name": "sdc", "type": "disk", "tran": null, "children": [
         {"name": "sdc1", "type": "part", "tran": null},
         {"name": "sdc2", "type": "part", "tran": null}
      ]},
      {"name": "sdg", "type": "disk", "tran": "fc", "children": [
         {"name": "36000d31005918c0000000000000004a6", "type": "mpath", "tran": null}
      ]},
      {"name": "sdr", "type": "disk", "tran": "fc", "children": [
         {"name": "sdr1", "type": "part", "tran": null},
         {"name": "36000d31005918c0000000000000004a5", "type": "mpath", "tran": null, "children": [
            {"name": "36000d31005918c0000000000000004a5p1", "type": "part", "tran": null}
         ]}
      ]},
      {"name": "sdae", "type": "disk", "tran": "fc", "children": [
         {"name": "36000d31005918c0000000000000004a6", "type": "mpath", "tran": null}
      ]},
      {"name": "sdap", "type": "disk", "tran": "fc", "children": [
         {"name": "sdap1", "type": "part", "tran": null},
         {"name": "36000d31005918c0000000000000004a5", "type": "mpath", "tran": null, "children": [
            {"name": "36000d31005918c0000000000000004a5p1", "type": "part", "tran": null}
         ]}
      ]}
   ]
}`

func parseLsblkTree(t *testing.T, out string) []lsblkNode {
	t.Helper()

	resp := map[string][]lsblkNode{}
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	return resp["blockdevices"]
}

func TestNonDasDisksFcMultipath(t *testing.T) {
	nonDas := nonDasDisks(parseLsblkTree(t, lsblkTreeWithFcMultipath))

	require.Equal(t, map[string]bool{"sdg": true, "sdr": true, "sdae": true, "sdap": true}, nonDas)
}

func TestNonDasDisksByTransportOrMultipath(t *testing.T) {
	devs := []lsblkNode{
		{Name: "sda", Type: "disk", Tran: "sata"},
		{Name: "nvme0n1", Type: "disk", Tran: "nvme"},
		{Name: "sdb", Type: "disk", Tran: "iscsi"},
		{Name: "sdc", Type: "disk", Tran: "sas", Children: []lsblkNode{{Name: "mpatha", Type: "mpath"}}},
		{Name: "sdd", Type: "disk", Tran: "sas", Children: []lsblkNode{
			{Name: "sdd1", Type: "part", Children: []lsblkNode{{Name: "mpathb", Type: "mpath"}}},
		}},
		{Name: "sr0", Type: "rom", Tran: "sata"},
	}

	require.Equal(t, map[string]bool{"sdb": true, "sdc": true, "sdd": true}, nonDasDisks(devs))
}

func TestGetBlockOrPartitionOnlySkipsNonDasDisks(t *testing.T) {
	raws := []nodes.RawBlockDevice{
		{Type: "disk", Name: "sda"},
		{Type: "part", Name: "sda1"},
		{Type: "disk", Name: "sdg"},
		{Type: "mpath", Name: "36000d31005918c0000000000000004a6"},
		{Type: "rom", Name: "sr0"},
	}

	got := getBlockOrPartitionOnly(raws, map[string]bool{"sdg": true})

	names := []string{}
	for _, raw := range got {
		names = append(names, raw.Name)
	}
	require.Equal(t, []string{"sda", "sda1"}, names)
}
