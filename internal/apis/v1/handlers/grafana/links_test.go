package grafana

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceDashboardLink(t *testing.T) {
	link := InstanceDashboardLink("vm-1")

	require.Contains(t, link, "/grafana/d/PVW6vU7Wz/instance")
	require.True(t, strings.HasSuffix(link, "var-UUID=vm-1"))
}

func TestInstanceVgpuHistoryLinks(t *testing.T) {
	workload := InstanceVgpuWorkloadHistoryLink("vm-1", "proj-1", "instance-1")
	vram := InstanceVgpuVramHistoryLink("vm-1", "proj-1", "instance-1")

	for _, link := range []string{workload, vram} {
		require.Contains(t, link, "/grafana/d/PVW6vU7Wz/instance")
		// Pinning all three is what keeps the page from labelling this VM's chart
		// with the tenant's alphabetically first VM.
		require.Contains(t, link, "var-TID=proj-1")
		require.Contains(t, link, "var-HOSTNAME=instance-1")
		require.Contains(t, link, "var-UUID=vm-1")
		// kiosk=tv was removed in Grafana 11; it must not be carried over.
		require.NotContains(t, link, "kiosk")
	}
	// The vGPU row ships collapsed, so each link has to open its own panel
	// directly: 36 is GPU Usage, 37 is Memory Usage.
	require.Contains(t, workload, "viewPanel=36")
	require.Contains(t, vram, "viewPanel=37")
}

// A VM name is user-supplied: unescaped, a space or '&' would truncate or
// corrupt the variable that follows it.
func TestInstanceVgpuHistoryLinkEscapesValues(t *testing.T) {
	link := InstanceVgpuWorkloadHistoryLink("vm-1", "proj-1", "my vm&x")

	require.Contains(t, link, "var-HOSTNAME=my+vm%26x")
}

// A link offered from one GPU's row has to name that card, or every row on the
// node yields the same chart.
func TestDeviceGpuHistoryLinksPinTheCard(t *testing.T) {
	util := DeviceGpuUtilizationHistoryLink("cn13", "00000001:04:00.0")
	vram := DeviceGpuVramHistoryLink("cn13", "00000001:04:00.0")

	for _, link := range []string{util, vram} {
		require.Contains(t, link, "/grafana/d/i-device/device")
		require.Contains(t, link, "var-GPU_HOST=cn13")
		// Colons have to survive as an escaped query value.
		require.Contains(t, link, "var-GPU_PCIID=00000001%3A04%3A00.0")
	}
	require.Contains(t, util, "viewPanel=50")
	require.Contains(t, vram, "viewPanel=51")
}

// An empty PCI address must drop the variable rather than send it empty: it
// defaults to All, while an empty value would match no card at all.
func TestDeviceGpuHistoryLinkOmitsEmptyCard(t *testing.T) {
	link := DeviceGpuUtilizationHistoryLink("cn13", "")

	require.Contains(t, link, "var-GPU_HOST=cn13")
	require.NotContains(t, link, "GPU_PCIID")
}
