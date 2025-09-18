package fixpacks

import (
	"sort"
)

func (h *helper) sortNodesByHost(nodes *[]node) {
	sort.Slice(*nodes, func(i, j int) bool {
		return (*nodes)[i].Name < (*nodes)[j].Name
	})
}
