package images

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
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
