package healths

import (
	"fmt"
	nativehttp "net/http"
	"net/url"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) reportToController(health *cubecos.Health) {
	token, err := genControllerAuthToken()
	if err != nil {
		log.Errorf("failed to gen controller auth token: %s", err.Error())
		return
	}

	url, err := genControllerHealthReportUrl()
	if err != nil {
		log.Errorf("failed to gen health report url: %s", err.Error())
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().SetCookie(token).SetBody(health).Put(url)
	if err != nil {
		log.Errorf("failed to report health: %s(%s)", err.Error(), resp.String())
	}
}

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
