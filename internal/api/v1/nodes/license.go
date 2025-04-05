package nodes

import (
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (h *helper) addLicenseToNode(node *definition.Node) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to add license info to the nodes: %s", api.GetReqId(h.c), err.Error())
		return
	}

	node.License = h.getLicenseByHostname(
		licenses,
		node.Hostname,
	)
}

func (h *helper) addLicenseInfoToNodes(nodes *[]definition.Node) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to add license info to the nodes: %s", api.GetReqId(h.c), err.Error())
		return
	}

	for i, node := range *nodes {
		(*nodes)[i].License = h.getLicenseByHostname(
			licenses,
			node.Hostname,
		)
	}
}

func (h *helper) getLicenseByHostname(licenses []definition.License, hostname string) definition.License {
	for _, license := range licenses {
		if slices.Contains(license.Hosts, hostname) {
			license.Hosts = nil
			return license
		}
	}

	return definition.License{}
}
