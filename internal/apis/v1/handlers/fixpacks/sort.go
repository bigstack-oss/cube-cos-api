package fixpacks

import (
	"sort"
)

func (h *helper) sortUpdatableNodesByHost(nodes *[]node) {
	sort.Slice(*nodes, func(i, j int) bool {
		return (*nodes)[i].Name < (*nodes)[j].Name
	})
}
