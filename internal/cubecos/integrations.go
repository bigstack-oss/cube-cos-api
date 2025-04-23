package cubecos

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func ListBuiltInIntegrations() []v1.Integration {
	return []v1.Integration{
		{
			Name:                    "keycloak",
			IsHeaderShortcutEnabled: true,
			Description:             "Single sign-on authentication and authorization service",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443/auth/admin", v1.DataCenterVip),
		},
		{
			Name:                    "openstack",
			IsHeaderShortcutEnabled: true,
			Description:             "Free and open-source cloud computing platform",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:9999/base/overview", v1.DataCenterVip),
		},
		{
			Name:                    "rancher",
			IsHeaderShortcutEnabled: true,
			Description:             "Turnkey Kubernetes container management platform",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443", v1.DataCenterVip),
		},
		{
			Name:                    "ceph",
			IsHeaderShortcutEnabled: true,
			Description:             "Software-defined storage platform built on a general-purpose distributed framework, supporting object storage, block storage, and file storage.",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:7443/ceph/#/dashboard", v1.DataCenterVip),
		},
	}
}
