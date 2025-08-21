package fixpacks

import (
	"sort"
)

func (h *helper) sortUpdatableNodes(nodes *[]node) {
	sort.Slice(*nodes, func(i, j int) bool {
		return (*nodes)[i].UpdatedAt > (*nodes)[j].UpdatedAt
	})
}
