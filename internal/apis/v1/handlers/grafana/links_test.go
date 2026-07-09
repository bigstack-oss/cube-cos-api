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
