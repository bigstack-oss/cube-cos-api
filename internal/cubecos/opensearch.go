package cubecos

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func GetOpenSearchRequestLink(requestId string) (string, error) {
	out, err := exec.Command("hex_sdk", "opensearch_ops_reqid_url", requestId).CombinedOutput()
	if err != nil {
		return "", err
	}

	if !IsHexSdkSuccess(err) {
		return "", fmt.Errorf("failed to get openserach request link from cos: %s", string(out))
	}

	link := string(out)
	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("failed to parse opensearch request link: %s (%v)", link, err)
	}

	segments := strings.Split(u.Host, ":")
	if len(segments) < 1 {
		return "", fmt.Errorf("invalid opensearch request link: %s", link)
	}

	segments[0] = base.DataCenterVip
	u.Host = strings.Join(segments, ":")
	return url.QueryUnescape(u.String())
}
