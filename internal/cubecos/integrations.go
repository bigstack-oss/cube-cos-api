package cubecos

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func ListBuiltInIntegrations() []definition.Integration {
	return []definition.Integration{
		{
			Name:                    "keycloak",
			IsHeaderShortcutEnabled: true,
			Description:             "Keycloak Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443/auth/admin", definition.DataCenterVip),
		},
		{
			Name:                    "openstack",
			IsHeaderShortcutEnabled: true,
			Description:             "OpenStack Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s/horizon", definition.DataCenterVip),
		},
		{
			Name:                    "rancher",
			IsHeaderShortcutEnabled: true,
			Description:             "Rancher Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:10443", definition.DataCenterVip),
		},
		{
			Name:                    "ceph",
			IsHeaderShortcutEnabled: true,
			Description:             "Ceph Dashboard",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:7443/ceph/#/dashboard", definition.DataCenterVip),
		},
	}
}
