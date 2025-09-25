package images

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumetypes"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
	log "go-micro.dev/v5/logger"
)

func (h *helper) listProjects() ([]Project, error) {
	openstack := openstack.GetGlobalHelper()
	opsProjects, err := openstack.ListProjects(&projects.ListOpts{})
	if err != nil {
		log.Errorf("images(%s): failed to list projects:(%v)", h.reqId, err)
		return nil, err
	}

	projects := []Project{}
	for _, opsProject := range opsProjects {
		projects = append(projects, Project{
			Name:        opsProject.Name,
			Domain:      opsProject.DomainID,
			Enabled:     opsProject.Enabled,
			Description: opsProject.Description,
		})
	}

	return projects, nil
}

func (h *helper) listDomains() ([]string, error) {
	openstack := openstack.GetGlobalHelper()
	opsDomains, err := openstack.ListDomains(&domains.ListOpts{})
	if err != nil {
		log.Errorf("images(%s): failed to list domains:(%v)", h.reqId, err)
		return nil, err
	}

	domains := []string{}
	for _, opsDomain := range opsDomains {
		domains = append(
			domains, opsDomain.Name,
		)
	}

	return domains, nil
}

func (h *helper) listDestinations() ([]destination, error) {
	types, err := h.openstack.ListVolumeTypes(volumetypes.ListOpts{})
	if err != nil {
		log.Errorf("images(%s): failed to list volume types:(%v)", h.reqId, err)
		return nil, err
	}

	defaultVolumeType, err := cubecos.GetDefaultVolumeType()
	if err != nil {
		log.Errorf("images(%s): failed to get default volume type:(%v)", h.reqId, err)
		return nil, err
	}

	destinations := []destination{}
	for _, t := range types {
		if t.Name == storages.DefaultType {
			continue
		}

		if !t.IsPublic {
			continue
		}

		destinations = append(destinations, destination{
			Name:      t.Name,
			IsDefault: t.Name == defaultVolumeType,
		})
	}

	return destinations, nil
}
