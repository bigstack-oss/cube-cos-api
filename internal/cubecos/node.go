package cubecos

import (
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	log "go-micro.dev/v5/logger"
)

func ListNodes() ([]v1.Node, error) {
	nodes := []v1.Node{}
	for _, role := range v1.Roles {
		roleNodes, err := v1.GetNodesByRole(role)
		if err != nil {
			log.Warnf("failed to get %s nodes: %s", role, err.Error())
			continue
		}
		if len(roleNodes) == 0 {
			continue
		}

		nodes = append(nodes, roleNodes...)
	}

	return nodes, nil
}

func ListHypervisors() ([]hypervisors.Hypervisor, error) {
	h := openstack.GetGlobalHelper()
	return h.ListHypervisors(hypervisors.ListOpts{})
}
