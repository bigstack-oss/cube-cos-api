package datacenters

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

func getNodeLicenseStatus() v1.NodeLicenseStatus {
	nodeStatus := v1.NodeLicenseStatus{}
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
