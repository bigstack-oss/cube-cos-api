package tunings

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func listNodesBySelector(nodes []*definition.Node, selector definition.Selector) []*definition.Node {
	if !selector.Enabled {
		return nodes
	}

	filteredNodes := []*definition.Node{}
	for _, node := range nodes {
		for key, value := range selector.Labels {
			if node.Labels[key] == value {
				filteredNodes = append(filteredNodes, node)
				break
			}
		}
	}

	return filteredNodes
}

func selectRolesUsingActivityAndLabels(tuningSpec *definition.TuningSpec) []*definition.Role {
	for i, role := range tuningSpec.Roles {
		tuningSpec.Roles[i].Nodes = listNodesBySelector(role.Nodes, tuningSpec.Selector)
	}

	roles := []*definition.Role{}
	for _, role := range tuningSpec.Roles {
		if !role.IsNodeEmpty() {
			roles = append(roles, role)
		}
	}

	return roles
}

func (h *helper) filterTunings(tunings []definition.Tuning) []definition.Tuning {
	result, err := h.searchTunings(tunings)
	if err != nil {
		log.Errorf("failed to search tunings: %s", err.Error())
		return nil
	}

	tuningMap := genTuningMap(tunings)
	filtered := []definition.Tuning{}
	for _, hit := range result.Hits {
		filtered = append(filtered, tuningMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchTunings(tunings []definition.Tuning) (*bleve.SearchResult, error) {
	searcher := definition.GetTuningSearcher()
	for _, tuning := range tunings {
		err := searcher.Index(tuning.Name, tuning)
		if err != nil {
			continue
		}
	}

	return searcher.Search(
		bleve.NewSearchRequestOptions(
			bleve.NewMatchQuery(h.keyword),
			1000,
			0,
			false,
		),
	)
}

func genTuningMap(tunings []definition.Tuning) map[string]definition.Tuning {
	tuningMap := map[string]definition.Tuning{}
	for _, tuning := range tunings {
		tuningMap[tuning.Name] = tuning
	}

	return tuningMap
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}
