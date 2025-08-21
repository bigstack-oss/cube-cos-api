package fixpacks

import (
	"fmt"
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterUnsupportedNodes(nodes []node) ([]node, error) {
	fixpack, found := cubecos.GetFixpackByVersion(h.version)
	if !found {
		err := fmt.Errorf("fixpack version %s not found", h.version)
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

	return supported, nil
}
