package fixpacks

import (
	"fmt"
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterNodesByRole(roles []string) ([]nodes.Node, error) {
	list := nodes.List()
	if len(list) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	nodes := []nodes.Node{}
	for _, node := range list {
		if slices.Contains(roles, node.Role) {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (h *helper) filterUnsupportedNodes(nodes []node, version string) ([]node, error) {
	fixpack, found := cubecos.GetFixpackRawByVersion(version)
	if !found {
		err := fmt.Errorf("fixpack version %s not found", version)
		log.Errorf("fixpack(%s): %s", h.reqId, err)
		return nil, err
	}

	if len(fixpack.SupportedFirmwares) == 0 {
		return nodes, nil
	}

	supported := make([]node, 0, len(nodes))
	for _, node := range nodes {
		if slices.Contains(fixpack.SupportedFirmwares, node.Version) {
			supported = append(supported, node)
		}
	}

	if len(supported) == 0 {
		return nil, fmt.Errorf(
			"no nodes support fixpack version %s",
			version,
		)
	}

	return supported, nil
}
