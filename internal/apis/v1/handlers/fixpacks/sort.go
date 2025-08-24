package fixpacks

import (
	"sort"
)

func (h *helper) sortUpdatableNodes(nodes *[]node) {
	sort.Slice(*nodes, func(i, j int) bool {
		return (*nodes)[i].UpdatedAt > (*nodes)[j].UpdatedAt
	})
}

func (h *helper) sortProgress(progress *[]progress) {
	sort.Slice(*progress, func(i, j int) bool {
		return (*progress)[i].Host < (*progress)[j].Host
	})
}
