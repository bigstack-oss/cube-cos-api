package cubecos

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

// note:
// the description of each integration is from https://www.bigstack.co/products/cubecos/feature
// currently, cos has no the source text for it, but UI needs it, so we can only place it below for the time being.
//
// also, cos is a bit hard to have a solid convention to fetch the port or path for the services,
// so we just hardcode the info here, but in the M2, can consider to discuss with team to support such features from cos.
func ListBuiltInIntegrations() []v1.Integration {
	return []v1.Integration{
		{
			Name:                    "keycloak",
			IsHeaderShortcutEnabled: true,
			Description:             "Single sign-on authentication and authorization service",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443/auth/admin", base.DataCenterVip),
		},
		{
			Name:                    "openstack",
			IsHeaderShortcutEnabled: true,
			Description:             "Free and open-source cloud computing platform",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:9999/base/overview", base.DataCenterVip),
		},
		{
			Name:                    "rancher",
			IsHeaderShortcutEnabled: true,
			Description:             "Turnkey Kubernetes container management platform",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443", base.DataCenterVip),
		},
		{
			Name:                    "ceph",
			IsHeaderShortcutEnabled: true,
			Description:             "Software-defined storage platform built on a general-purpose distributed framework, supporting object storage, block storage, and file storage.",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:7443/ceph/#/dashboard", base.DataCenterVip),
		},
	}
}
