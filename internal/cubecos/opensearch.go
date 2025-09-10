package cubecos

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetOpenSearchRequestLink(requestId string) (string, error) {
	out, err := exec.Command("hex_sdk", "opensearch_ops_reqid_url", requestId).CombinedOutput()
	if err != nil {
		return "", err
	}

	if !IsHexSuccessful(err) {
		return "", fmt.Errorf("failed to get openserach request link from cos: %s", string(out))
	}

	link := string(out)
	return strings.Trim(link, "\n"), nil
}
