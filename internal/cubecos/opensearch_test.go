package cubecos

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func TestPublicOpenSearchLink(t *testing.T) {
	fqdn := base.DataCenterFqdn
	t.Cleanup(func() { base.DataCenterFqdn = fqdn })

	link := "http://10.1.0.100:5601/opensearch-dashboards/app/data-explorer/discover/#/view/req-1?_a=(x:!(a,b))"

	base.DataCenterFqdn = ""
	if got := publicOpenSearchLink(link); got != link {
		t.Errorf("without an fqdn the link changed: %s", got)
	}

	base.DataCenterFqdn = "cos.example.test"
	want := "https://cos.example.test/opensearch-dashboards/app/data-explorer/discover/#/view/req-1?_a=(x:!(a,b))"
	if got := publicOpenSearchLink(link); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	if got := publicOpenSearchLink("not a link"); got != "not a link" {
		t.Errorf("an unparsable link changed: %s", got)
	}
}
