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
			Description:             "Keycloak Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443/auth/admin", v1.DataCenterVip),
		},
		{
			Name:                    "openstack",
			IsHeaderShortcutEnabled: true,
			Description:             "OpenStack Skyline Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:9999/base/overview", v1.DataCenterVip),
		},
		{
			Name:                    "rancher",
			IsHeaderShortcutEnabled: true,
			Description:             "Rancher Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443", v1.DataCenterVip),
		},
		{
			Name:                    "ceph",
			IsHeaderShortcutEnabled: true,
			Description:             "Ceph Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:7443/ceph/#/dashboard", v1.DataCenterVip),
		},
	}
}
