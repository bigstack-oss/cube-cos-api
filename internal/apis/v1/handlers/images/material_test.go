package images

import (
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/stretchr/testify/require"
)

func TestVisibleDomainNamesHidesServiceDomains(t *testing.T) {
	opsDomains := []domains.Domain{
		{Name: "Default"},
		{Name: "heat"},
		{Name: "customer"},
	}

	require.Equal(t, []string{"Default", "customer"}, visibleDomainNames(opsDomains))
}

func TestVisibleDomainNamesKeepsLookalikeDomains(t *testing.T) {
	opsDomains := []domains.Domain{
		{Name: "Heat"},
		{Name: "heat-team"},
	}

	require.Equal(t, []string{"Heat", "heat-team"}, visibleDomainNames(opsDomains))
}

func TestVisibleDomainNamesEmpty(t *testing.T) {
	require.Equal(t, []string{}, visibleDomainNames(nil))
}
