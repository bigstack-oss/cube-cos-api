package v1

import "fmt"

const (
	RadosGatewayPort = 8888
)

func GetRadosGatewayUrl() string {
	return fmt.Sprintf("http://%s:%d/", DataCenterVip, RadosGatewayPort)
}
