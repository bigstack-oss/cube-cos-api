package datacenters

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

func getLicenseStatus() v1.LicenseStatus {
	nodes := v1.ListNodes()
	licenseStatus := v1.LicenseStatus{}

	for _, node := range nodes {
		switch node.License.Status.Current {
		case status.Valid:
			licenseStatus.Valid++
		case status.Expired:
			licenseStatus.Expired++
		case status.Unlicense:
			licenseStatus.Unlicensed++
		}
	}

	return licenseStatus
}
