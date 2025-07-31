package integration

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module = "integrations"
)

// note:
// the description of each integration is from https://www.bigstack.co/products/cubecos/feature
// currently, cos has no the source text for it, but UI needs it, so we can only place it below for the time being.
//
// also, cos is a bit hard to have a solid convention to fetch the port or path for the services,
// so we just hardcode the info here, but in the M2, can consider to discuss with team to support such features from cos.
type Service struct {
	Name                    string `json:"name"`
	IsHeaderShortcutEnabled bool   `json:"isHeaderShortcutEnabled"`
	Description             string `json:"description"`
	IsBuiltIn               bool   `json:"isBuiltIn"`
	Url                     string `json:"url"`
}

//	{
//	    "name": "cubeStorage",
//	    "type": "built-in",
//	    "vendor": "testintegration",
//	    "managementIp": "",
//	    "updatedAt": "2023-10-01T12:00:00+08:00",
//	    "isDefault": true,
//	    "status": {
//	        "current": "active",
//	        "isProcessing": false
//	    }
//	}
type Storage struct {
	Name         string             `json:"name"`
	Type         string             `json:"type"` // built-in, third-party
	Vendor       string             `json:"vendor"`
	ManagementIp string             `json:"managementIp"`
	UpdatedAt    string             `json:"updatedAt"`
	IsDefault    bool               `json:"isDefault"`
	Status       status.Integration `json:"status"`
}

func GetCommonServices() []Service {
	return []Service{
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
			Name:                    "ceph",
			IsHeaderShortcutEnabled: true,
			Description:             "Software-defined storage platform built on a general-purpose distributed framework, supporting object storage, block storage, and file storage.",
			IsBuiltIn:               true,
			Url:                     fmt.Sprintf("https://%s:7443/ceph/#/dashboard", base.DataCenterVip),
		},
	}
}

func GetCloudService() Service {
	return Service{
		Name:                    "rancher",
		IsHeaderShortcutEnabled: true,
		Description:             "Turnkey Kubernetes container management platform",
		IsBuiltIn:               true,
		Url:                     fmt.Sprintf("https://%s:10443", base.DataCenterVip),
	}
}
