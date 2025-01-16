package cubecos

import (
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
)

func ListNodes() ([]*definition.Node, error) {
	nodes := []*definition.Node{}

	for _, role := range definition.Roles {
		ns, err := definition.GetNodesByRole(role)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, ns...)
	}

	return nodes, nil
}

func ListHypervisors() ([]hypervisors.Hypervisor, error) {
	h, err := openstack.NewHelper()
	if err != nil {
		return nil, err
	}

	return h.ListHypervisors(hypervisors.ListOpts{})
}
