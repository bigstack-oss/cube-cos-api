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

func TestInstanceVgpuDashboardLink(t *testing.T) {
	link := InstanceVgpuDashboardLink("vm-1", "proj-1", "instance-1")

	require.Contains(t, link, "/grafana/d/PVW6vU7Wz/instance")
	// Pinning all three is what keeps the page from labelling this VM's chart
	// with the tenant's alphabetically first VM.
	require.Contains(t, link, "var-TID=proj-1")
	require.Contains(t, link, "var-HOSTNAME=instance-1")
	require.Contains(t, link, "var-UUID=vm-1")
	// The vGPU row ships collapsed, so the link has to open a panel directly.
	require.Contains(t, link, "viewPanel=37")
	// kiosk=tv was removed in Grafana 11; it must not be carried over.
	require.NotContains(t, link, "kiosk")
}

// A VM name is user-supplied: unescaped, a space or '&' would truncate or
// corrupt the variable that follows it.
func TestInstanceVgpuDashboardLinkEscapesValues(t *testing.T) {
	link := InstanceVgpuDashboardLink("vm-1", "proj-1", "my vm&x")

	require.Contains(t, link, "var-HOSTNAME=my+vm%26x")
}
