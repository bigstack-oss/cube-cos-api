package nodes

import (
	"fmt"
	"net/url"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (n *Node) GenUrl() string {
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
	u.RawQuery = "clusterWise=false"
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

func (n *Node) PatchTriggerTaskUrl(trigger triggers.ApiSchema) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/tasks/%s", base.DataCenterName, trigger.Name)
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
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/settings/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) CreateDeviceUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) PatchDeviceTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/devices/tasks", base.DataCenterName, n.Hostname)
	return u.String()
}

func (n *Node) PatchOsdTaskUrl() string {
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
