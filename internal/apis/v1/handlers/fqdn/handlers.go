package fqdn

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
)

// Serving the web UIs on an FQDN (cubesys.fqdn).
//
// The CLI on a control node (cluster> fqdn) calls the first four with the node token. The
// control it reaches applies the change on itself, then on every other node one at a time
// through the /fqdn/node endpoints, then runs the cluster-wide part once: Keycloak's hostname,
// the SAML clients, Rancher, the ceph dashboard. Each node keeps the FQDN as a hex_config
// tuning, so it is applied again when the node bootstraps.
var Handlers = []apis.Handler{
	{
		Version: apis.V1,
		Method:  http.MethodGet,
		Path:    "/fqdn",
		Func:    getFqdn,
	},
	{
		Version: apis.V1,
		Method:  http.MethodPut,
		Path:    "/fqdn",
		Func:    setFqdn,
	},
	{
		Version: apis.V1,
		Method:  http.MethodDelete,
		Path:    "/fqdn",
		Func:    resetFqdn,
	},
	{
		Version: apis.V1,
		Method:  http.MethodPut,
		Path:    "/fqdn/certificate",
		Func:    installCertificate,
	},
	{
		Version: apis.V1,
		Method:  http.MethodGet,
		Path:    "/fqdn/node",
		Func:    getNode,
	},
	{
		Version: apis.V1,
		Method:  http.MethodPut,
		Path:    "/fqdn/node",
		Func:    setNode,
	},
	{
		Version: apis.V1,
		Method:  http.MethodPut,
		Path:    "/fqdn/node/certificate",
		Func:    reloadNodeCertificate,
	},
	{
		Version: apis.V1,
		Method:  http.MethodPut,
		Path:    "/fqdn/node/api",
		Func:    restartNodeApi,
	},
}
