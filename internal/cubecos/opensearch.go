package cubecos

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func GetOpenSearchRequestLink(requestId string) (string, error) {
	out, err := exec.Command("hex_sdk", "opensearch_ops_reqid_url", requestId).CombinedOutput()
	if err != nil {
		return "", err
	}

	if !IsHexSuccessful(err) {
		return "", fmt.Errorf("failed to get openserach request link from cos: %s", string(out))
	}

	link := strings.Trim(string(out), "\n")
	return publicOpenSearchLink(link), nil
}

// publicOpenSearchLink moves a link from hex_sdk, which points at OpenSearch Dashboards on
// http://<vip>:5601, onto the FQDN once one is set: through haproxy on https, because Keycloak
// and the ceph dashboard send HSTS for the name and a browser rewrites http://<fqdn>:5601 to
// https, which that port does not serve.
func publicOpenSearchLink(link string) string {
	if base.DataCenterFqdn == "" {
		return link
	}

	scheme, rest, found := strings.Cut(link, "://")
	if !found || scheme == "" {
		return link
	}

	_, path, _ := strings.Cut(rest, "/")
	return fmt.Sprintf("https://%s/%s", base.DataCenterFqdn, path)
}
