package cubecos

import (
	openstack "github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

func GetVmStatusOverview() (*Vm, error) {
	h, err := openstack.NewHelper()
	if err != nil {
		return nil, err
	}

	servers, err := h.ListServers(servers.ListOpts{AllTenants: true})
	if err != nil {
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
