package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

func getNodeLicenseStatus() base.NodeLicenseStatus {
	nodeStatus := base.NodeLicenseStatus{}
	for _, node := range nodes.List() {
		switch node.License.Status.Current {
		case status.Valid:
			nodeStatus.Valid++
		case status.Expired:
			nodeStatus.Expired++
		default:
			nodeStatus.Unlicense++
		}
	}

	return nodeStatus
}
