package cubecos

import (
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	log "go-micro.dev/v5/logger"
)

func GetVmStatusOverview() (*Vm, error) {
	h, err := openstack.NewHelper(
		openstack.AuthType(config.Data.Spec.Openstack.Auth.Type),
		openstack.AuthUrl(config.Data.Spec.Openstack.Auth.Url),
		openstack.ProjectName(config.Data.Spec.Openstack.Auth.Project.Name),
		openstack.ProjectDomainName(config.Data.Spec.Openstack.Auth.Project.Domain.Name),
		openstack.Username(config.Data.Spec.Openstack.Auth.Username),
		openstack.Password(config.Data.Spec.Openstack.Auth.Password),
	)
	if err != nil {
		log.Errorf("failed to create openstack helper: %v", err)
		return nil, err
	}

	servers, err := h.ListServers(servers.ListOpts{AllTenants: true})
	if err != nil {
		log.Errorf("failed to list servers: %v", err)
		return nil, err
	}

	return genVmStatusOverview(servers), nil
}

func genVmStatusOverview(servers []servers.Server) *Vm {
	vm := &Vm{Total: len(servers)}

	for _, server := range servers {
		switch server.PowerState.String() {
		case "RUNNING":
			vm.Running++
		case "SHUTDOWN":
			vm.Stopped++
		case "SUSPENDED":
			vm.Suspend++
		case "PAUSED":
			vm.Paused++
		case "CRASHED":
			vm.Error++
		default:
			vm.Unknown++
		}
	}

	return vm
}
