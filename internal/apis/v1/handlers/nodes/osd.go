package nodes

import nodes "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"

func (h *helper) setOsdNotFoundInfo(blockDev *nodes.BlockDevice) {
	blockDev.Status.Description = "osd not found from ceph device list"
	blockDev.Osd = nodes.Osd{Daemons: []nodes.Deamon{}}
}
