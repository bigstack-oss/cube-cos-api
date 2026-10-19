package images

import (
	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumetypes"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	log "go-micro.dev/v5/logger"
)

// hiddenServiceDomains are Keystone domains CubeCOS creates for its own
// services (heat holds Heat's stack users and projects), not for tenants.
// They are left out of the domain choices, as sdk_os.sh does for the CLI.
var hiddenServiceDomains = []string{
	"heat",
}

func (h *helper) listProjects() ([]Project, error) {
	if !h.caller.isAdmin {
		return h.listCallerProject()
	}

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

func (h *helper) listCallerProject() ([]Project, error) {
	if !h.caller.hasProject() {
		return []Project{}, nil
	}

	opsProject, err := h.openstack.GetProject(h.caller.projectId)
	if err != nil {
		log.Errorf("images(%s): failed to get project %s:(%v)", h.reqId, h.caller.projectId, err)
		return nil, err
	}

	return []Project{
		{
			Name:        opsProject.Name,
			Domain:      opsProject.DomainID,
			Enabled:     opsProject.Enabled,
			Description: opsProject.Description,
		},
	}, nil
}

func (h *helper) listDomains() ([]string, error) {
	if !h.caller.isAdmin {
		return h.listCallerDomain(), nil
	}

	openstack := openstack.GetGlobalHelper()
	opsDomains, err := openstack.ListDomains(&domains.ListOpts{})
	if err != nil {
		log.Errorf("images(%s): failed to list domains:(%v)", h.reqId, err)
		return nil, err
	}

	return visibleDomainNames(opsDomains), nil
}

func visibleDomainNames(opsDomains []domains.Domain) []string {
	names := []string{}
	for _, opsDomain := range opsDomains {
		if slices.Contains(hiddenServiceDomains, opsDomain.Name) {
			continue
		}

		names = append(names, opsDomain.Name)
	}

	return names
}

func (h *helper) listCallerDomain() []string {
	if !h.caller.hasProject() {
		return []string{}
	}

	return []string{h.caller.domainName}
}

// Reserved images and public visibility are offered to admins only, as only
// admins may import them.
func (h *helper) listReservedImages() []images.ReqOpts {
	if !h.caller.isAdmin {
		return []images.ReqOpts{}
	}

	return images.GetReserved()
}

func (h *helper) listVisibilities() []string {
	if !h.caller.isAdmin {
		return []string{string(opsimage.ImageVisibilityPrivate)}
	}

	return images.Visibilitise
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
