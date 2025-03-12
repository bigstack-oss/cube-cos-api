package tunings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func selectRolesUsingActivityAndLabels(tuningSpec *definition.TuningSpec) []*definition.Role {
	for i, role := range tuningSpec.Roles {
		tuningSpec.Roles[i].Nodes = getNodesBySelector(role.Nodes, tuningSpec.Selector)
	}

	roles := []*definition.Role{}
	for _, role := range tuningSpec.Roles {
		if !role.IsNodeEmpty() {
			roles = append(roles, role)
		}
	}

	return roles
}

func getNodesBySelector(nodes []*definition.Node, selector definition.Selector) []*definition.Node {
	if !selector.Enabled {
		return nodes
	}

	filtered := []*definition.Node{}
	for _, node := range nodes {
		for key, value := range selector.Labels {
			if node.Labels[key] == value {
				filtered = append(filtered, node)
				break
			}
		}
	}

	return filtered
}

func (h *helper) filterTunings(tunings []definition.Tuning) []definition.Tuning {
	if !h.isFilterRequired() {
		return tunings
	}

	if h.isKeywordRequired() {
		tunings = h.filteredByKeyword(tunings)
	}

	if h.isHostsRequired() {
		tunings = h.filteredByHosts(tunings)
	}

	if h.isModifiedRequired() {
		tunings = h.filteredByModified(tunings)
	}

	return tunings
}

func (h *helper) filteredByKeyword(tunings []definition.Tuning) []definition.Tuning {
	result, err := h.searchTunings(tunings)
	if err != nil {
		log.Errorf("failed to search tunings: %s", err.Error())
		return tunings
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
			maxSearchResults,
			0,
			false,
		),
	)
}

func (h *helper) filteredByHosts(tunings []definition.Tuning) []definition.Tuning {
	filtered := []definition.Tuning{}
	for _, tuning := range tunings {
		if h.containsHosts(tuning.Hosts) {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}

func (h *helper) filteredByModified(tunings []definition.Tuning) []definition.Tuning {
	filtered := []definition.Tuning{}
	for _, tuning := range tunings {
		if tuning.IsModified {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}

func (h *helper) containsHosts(hosts []definition.Host) bool {
	hostSet := make(map[string]struct{}, len(hosts))
	for _, h := range hosts {
		hostSet[h.Name] = struct{}{}
	}

	for _, h := range h.hosts {
		_, found := hostSet[h]
		if !found {
			return false
		}
	}

	return true
}

func genTuningMap(tunings []definition.Tuning) map[string]definition.Tuning {
	tuningMap := map[string]definition.Tuning{}
	for _, tuning := range tunings {
		tuningMap[tuning.Name] = tuning
	}

	return tuningMap
}

func (h *helper) enrichTuningPayload(tunings *[]definition.Tuning) {
	h.syncUpdates(tunings)
	h.sortTunings(tunings)
}

func (h *helper) syncUpdates(tunings *[]definition.Tuning) {
	updatingTunings, err := getUpdatingTunings()
	if err != nil {
		log.Errorf("tunings(%s): failed to get updating tunings: %s", api.GetReqId(h.c), err.Error())
		return
	}

	updatingTuningMap := genUpdatingTuningMap(updatingTunings)
	for i, tuning := range *tunings {
		(*tunings)[i].Status = &status.Tuning{}
		updatingTuning, found := updatingTuningMap[tuning.GenerateId()]
		if found {
			(*tunings)[i].Status.IsUpdating = updatingTuning.Status.IsUpdating
			(*tunings)[i].Status.UpdatedAt = updatingTuning.Status.UpdatedAt
		}
	}
}

func genUpdatingTuningMap(updateTunings []definition.Tuning) map[string]definition.Tuning {
	updateTuningMap := map[string]definition.Tuning{}
	for _, tuning := range updateTunings {
		updateTuningMap[tuning.GenerateId()] = tuning
	}

	return updateTuningMap
}
