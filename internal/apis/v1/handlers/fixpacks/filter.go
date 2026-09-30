package fixpacks

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	log "go-micro.dev/v5/logger"
)

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

	supported := []node{}
	for _, node := range nodes {
		segment := strings.Split(node.Version, " ")
		if len(segment) < 3 {
			continue
		}

		numericVersion := segment[2]
		if slices.Contains(fixpack.SupportedFirmwares, numericVersion) {
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
