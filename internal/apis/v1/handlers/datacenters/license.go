package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

func getNodeLicenseStatus() base.NodeLicenseStatus {
	nodeStatus := base.NodeLicenseStatus{}
	nodes := cubecos.ListNodesWithTimeSensitiveInfo()

	for _, node := range nodes {
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
