package grafana

import (
	"fmt"
	"net/url"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/gin-gonic/gin"
)

func genHostLink(c *gin.Context) string {
	return fmt.Sprintf(
		"https://%s/grafana/d/i-R2q81iz/host?refresh=5m&kiosk=tv&orgId=1&var-HOST=%s",
		base.DataCenterVip,
		c.Param("hostname"),
	)
}

// InstanceDashboardLink is the Grafana instance dashboard URL for a VM, as
// served by the generic /grafana/instances/{instanceId} endpoint. That endpoint
// knows nothing but the id, so the link cannot pin the dashboard's tenant and
// hostname variables; callers that do know them should use
// InstanceVgpuDashboardLink instead.
func InstanceDashboardLink(instanceId string) string {
	return fmt.Sprintf(
		"https://%s/grafana/d/PVW6vU7Wz/instance?refresh=5m&kiosk=tv&orgId=1&var-UUID=%s",
		base.DataCenterVip,
		instanceId,
	)
}

// instanceVgpuMemoryPanelId is the Memory Usage panel in the instance
// dashboard's vGPU row (dashboard UID PVW6vU7Wz). Pinned here as a deep-link
// target, so renumbering or reordering that row's panels in instance.json
// breaks these links -- same cross-team contract as the device dashboard's
// panels 50 and 51.
const instanceVgpuMemoryPanelId = 37

// InstanceVgpuDashboardLink is the deep link to one VM's vGPU usage.
//
// Pinning var-UUID alone is not enough. UUID is the tail of a
// TID -> TENANT -> HOSTNAME -> UUID chain of query variables that all refresh on
// dashboard load, so leaving the first two unset lets them resolve to whatever
// their query returns first -- in practice the tenant's alphabetically first VM,
// which then labels the page with a different VM than the one being charted.
// var-TID is matched against the metrics' tenant_id tag and collapses TENANT to
// the owning project; var-HOSTNAME pins the picker to this VM.
//
// viewPanel is required because the vGPU row ships collapsed: a link to the
// dashboard alone lands on a page with no GPU chart visible. Memory is the panel
// to open because it renders for every vGPU flavour, while GPU utilization is
// empty for a MIG-backed vGPU by design.
func InstanceVgpuDashboardLink(instanceId, tenantId, vmName string) string {
	return fmt.Sprintf(
		"https://%s/grafana/d/PVW6vU7Wz/instance?orgId=1&refresh=5m&var-TID=%s&var-HOSTNAME=%s&var-UUID=%s&from=now-3h&to=now&viewPanel=%d",
		base.DataCenterVip,
		url.QueryEscape(tenantId),
		// A VM name is user-supplied and may contain spaces or '&'.
		url.QueryEscape(vmName),
		url.QueryEscape(instanceId),
		instanceVgpuMemoryPanelId,
	)
}

func genInstanceLink(c *gin.Context) string {
	return InstanceDashboardLink(c.Param("instanceId"))
}

func genTopHostLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/M3ncw6lmk/top-hosts?refresh=5m&kiosk=tv&orgId=1",
		base.DataCenterVip,
	)
}

func genTopInstanceLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/qzfq087Wk/top-instances?refresh=5m&orgId=1&var-TID=&var-TOP=50&var-TENANT=admin",
		base.DataCenterVip,
	)
}

func genNetworksLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/Xx2kkftWk/network?orgId=1&refresh=5m",
		base.DataCenterVip,
	)
}

func genNetworkDevicesLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/i-device/device?refresh=5m&orgId=1",
		base.DataCenterVip,
	)
}

func genStoragesLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/QTc_sAxiw/storage?refresh=5m&kiosk=tv&orgId=1",
		base.DataCenterVip,
	)
}

// The device dashboard's GPU panel ids, pinned here as deep-link targets:
// renumbering or reordering them in device.json breaks these links.
const (
	deviceGpuUtilizationPanelId = 50
	deviceGpuVramPanelId        = 51
)

// deviceGpuHistoryLink builds a device-dashboard deep link (UID i-device) to one
// GPU history panel.
//
// $GPU_HOST is gpu.host's `host` tag, NOT $HOST (which comes from ipmi_sensor and
// can differ). pciAddress is optional and scopes the chart to a single card via
// the hidden $GPU_PCIID variable, whose values are gpu.host's `pciid` tag --
// byte-identical to the pciAddress the GPU listing reports for the same card, so
// no translation is needed. Passing an empty pciAddress omits the variable
// entirely rather than sending it empty: it defaults to All, and an empty value
// would match no card at all.
func deviceGpuHistoryLink(hostname, pciAddress string, panelId int) string {
	link := fmt.Sprintf(
		"https://%s/grafana/d/i-device/device?orgId=1&var-GPU_HOST=%s",
		base.DataCenterVip,
		url.QueryEscape(hostname),
	)
	if pciAddress != "" {
		// A PCI address carries colons, so it has to be escaped to survive as a
		// query value.
		link += "&var-GPU_PCIID=" + url.QueryEscape(pciAddress)
	}

	return fmt.Sprintf("%s&from=now-3h&to=now&viewPanel=%d", link, panelId)
}

// DeviceGpuUtilizationHistoryLink and DeviceGpuVramHistoryLink are the
// per-card history links, exported for the GPU listing handler to report inline
// with each card. The UI offers them from a single GPU's row, so they have to
// answer for that card rather than for every card on the node.
func DeviceGpuUtilizationHistoryLink(hostname, pciAddress string) string {
	return deviceGpuHistoryLink(hostname, pciAddress, deviceGpuUtilizationPanelId)
}

func DeviceGpuVramHistoryLink(hostname, pciAddress string) string {
	return deviceGpuHistoryLink(hostname, pciAddress, deviceGpuVramPanelId)
}

// The node-scoped endpoints keep answering for the whole node: they are handed
// only a hostname, so they cannot name a card.
func genGpuUtilizationHistoryLink(c *gin.Context) string {
	return deviceGpuHistoryLink(c.Param("hostname"), "", deviceGpuUtilizationPanelId)
}

func genGpuVramHistoryLink(c *gin.Context) string {
	return deviceGpuHistoryLink(c.Param("hostname"), "", deviceGpuVramPanelId)
}
