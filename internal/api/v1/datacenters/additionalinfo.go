package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

func getNodeLicenseStatus() datacenters.NodeLicenseStatus {
	nodeStatus := datacenters.NodeLicenseStatus{}
	for _, node := range nodes.List() {
		switch node.License.Status.Current {
		case status.Valid:
			nodeStatus.Valid++
		case status.Expired:
			nodeStatus.Expired++
		case status.Unlicense:
			nodeStatus.Unlicense++
		}
	}

	return nodeStatus
}
