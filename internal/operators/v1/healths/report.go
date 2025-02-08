package healths

import (
	"fmt"
	nativehttp "net/http"
	"net/url"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func genControllerAuthToken() (*nativehttp.Cookie, error) {
	nodes, err := definition.GetControllerNodes()
	if err != nil {
		log.Errorf("failed to get controller nodes: %s", err.Error())
		return nil, err
	}

	return &nativehttp.Cookie{
		Name:  "Bearer",
		Value: nodes[0].Id,
	}, nil

}

func genControllerHealthReportUrl() (string, error) {
	nodes, err := definition.GetControllerNodes()
	if err != nil {
		log.Errorf("failed to get controller nodes: %s", err.Error())
		return "", err
	}

	u := url.URL{
		Scheme: "http",
		Host:   nodes[0].Address,
		Path:   fmt.Sprintf("/api/v1/datacenters/%s/healths/all", definition.DataCenterName),
	}

	return u.String(), nil
}
