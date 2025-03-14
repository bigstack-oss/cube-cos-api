package nodes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

func addLicenseToNode(c *gin.Context, node *definition.Node) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to add license info to the nodes: %s", api.GetReqId(c), err.Error())
		return
	}

	node.License = getLicenseByHostname(
		licenses,
		node.Hostname,
	)
}

func addLicenseInfoToNodes(c *gin.Context, nodes *[]*definition.Node) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to add license info to the nodes: %s", api.GetReqId(c), err.Error())
		return
	}

	for _, node := range *nodes {
		node.License = getLicenseByHostname(licenses, node.Hostname)
	}
}

func getLicenseByHostname(licenses []definition.License, hostname string) definition.License {
	for _, license := range licenses {
		if license.Hostname == hostname {
			return license
		}
	}

	return definition.License{}
}
