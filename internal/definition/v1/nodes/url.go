package nodes

import (
	"fmt"
	"net/url"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

func (n *Node) GenUrl() url.URL {
	return url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
	}
}

func (n *Node) GenUrlString() string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	return u.String()
}

func (n *Node) GetMetricUrl(metric, view string) string {
	u := url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
		Path: fmt.Sprintf(
			"/api/v1/datacenters/%s/metrics/%s/%s/hosts/%s",
			n.DataCenter,
			metric,
			view,
			n.Hostname,
		),
	}

	return u.String()
}

func (n *Node) GetNodeUrl() string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s", n.DataCenter, n.Hostname)
	return u.String()
}

func (n *Node) RebootNodeUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/softReboot", n.DataCenter, n.Hostname)
	return u.String()
}

func (n *Node) PostLicenseUrl() string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/licenses/hosts/%s", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) GetTuningUrl() string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters", n.DataCenter)
	u.RawQuery = "allNodes=false"
	return u.String()
}

func (n *Node) GetSettingUrl(path string) string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	u.Path = path
	return u.String()
}

func (n *Node) GetSupportFileUrl() string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/hosts/%s", n.DataCenter, n.Hostname)
	return u.String()
}

func (n *Node) DownloadSupportFileUrl(setname, filename string) string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/%s/%s", n.DataCenter, setname, filename)
	return u.String()
}

func (n *Node) PatchTuningUrl(tuning string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters/%s", base.DataCenterName, tuning)
	return u.String()
}

func (n *Node) EnableOrDisableTuningUrl(tuning string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters/%s/enable", base.DataCenterName, tuning)
	return u.String()
}

func (n *Node) ResetTuningUrl(tuning string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters/%s/reset", base.DataCenterName, tuning)
	return u.String()
}

func (n *Node) PatchTuningTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) PostTriggerUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers", base.DataCenterName)
	return u.String()
}

func (n *Node) UpdateTriggerUrl(trigger string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/%s", base.DataCenterName, trigger)
	return u.String()
}

func (n *Node) ToggleTriggerUrl(trigger string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/%s/enable", base.DataCenterName, trigger)
	return u.String()
}

func (n *Node) PatchTriggerTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) CreateSupportFileUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles", base.DataCenterName)
	return u.String()
}

func (n *Node) PatchSupportFileTaskUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/%s", base.DataCenterName, file.Group)
	return u.String()
}

func (n *Node) DeleteSupportFileUrl(group, file string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/%s/%s", base.DataCenterName, group, file)
	return u.String()
}

func (n *Node) PatchSettingTaskUrl(setting settings.Setting) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/settings/tasks/%s", base.DataCenterName, base.Hostname)
	return u.String()
}

func (n *Node) ListDevicesUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) GetDeviceUrl(device string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices/%s", base.DataCenterName, n.Hostname, device)
	return u.String()
}

func (n *Node) AddDeviceUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) UpdateDeviceUrl(device string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices/%s", base.DataCenterName, n.Hostname, device)
	return u.String()
}

func (n *Node) RemoveDeviceUrl(device string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices/%s", base.DataCenterName, n.Hostname, device)
	return u.String()
}

func (n *Node) UpdateDeviceTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices/tasks", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) GetOsdUrl(id string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/osds/%s", base.DataCenterName, n.Hostname, id)
	return u.String()
}

func (n *Node) RestartOsdUrl(id string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/osds/%s/restart", base.DataCenterName, n.Hostname, id)
	return u.String()
}

func (n *Node) PatchOsdUrl(id string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/osds/%s", base.DataCenterName, n.Hostname, id)
	return u.String()
}

func (n *Node) RemoveOsdUrl(id string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/osds/%s", base.DataCenterName, n.Hostname, id)
	return u.String()
}

func (n *Node) UpdateOsdTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/osds/tasks", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) DeleteRepairingTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/healths/tasks/repairing", base.DataCenterName)
	return u.String()
}

func (n *Node) DeleteModuleRepairingTaskUrl(module string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/healths/tasks/repairing/%s", base.DataCenterName, module)
	return u.String()
}

func (n *Node) UpdateImageTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/images/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) UpdateVolumeImageTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/volumes/images/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) UpdateFirmwareTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/firmwares/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) PatchFixpackUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/fixpacks", base.DataCenterName)
	return u.String()
}

func (n *Node) PostFixpackRollbackUrl(version string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/fixpacks/%s/rollback", base.DataCenterName, version)
	return u.String()
}

func (n *Node) UpdateFixpackTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/fixpacks/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) PostStorageUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages", base.DataCenterName)
	return u.String()
}

func (n *Node) PatchStorageUrl(name string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/%s", base.DataCenterName, name)
	return u.String()
}

func (n *Node) DeleteStorageUrl(name string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/%s", base.DataCenterName, name)
	return u.String()
}

func (n *Node) UpdateStorageTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) PostStorageModelUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/models", base.DataCenterName)
	return u.String()
}

func (n *Node) PatchStorageModelUrl(vendor, product string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/models/%s/%s", base.DataCenterName, vendor, product)
	return u.String()
}

func (n *Node) DeleteStorageModelUrl(vendor, product string) string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/models/%s/%s", base.DataCenterName, vendor, product)
	return u.String()
}

func (n *Node) UpdateModelTaskUrl() string {
	u := n.GenUrl()
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/integrations/storages/models/tasks", base.DataCenterName)
	return u.String()
}
