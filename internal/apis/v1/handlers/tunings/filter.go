package tunings

import (
	"slices"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func selectRolesUsingActivityAndLabels(tuningSpec *tunings.Spec) []*nodes.Role {
	for i, role := range tuningSpec.Roles {
		tuningSpec.Roles[i].Nodes = getNodesBySelector(role.Nodes, tuningSpec.Selector)
	}

	roles := []*nodes.Role{}
	for _, role := range tuningSpec.Roles {
		if !role.IsNodeEmpty() {
			roles = append(roles, role)
		}
	}

	return roles
}

func getNodesBySelector(nodesToFilter []nodes.Node, selector nodes.Selector) []nodes.Node {
	if !selector.Enabled {
		return nodesToFilter
	}

	filtered := []nodes.Node{}
	for _, node := range nodesToFilter {
		for key, value := range selector.Labels {
			if node.Labels[key] == value {
				filtered = append(filtered, node)
				break
			}
		}
	}

	return filtered
}

func (h *helper) filterTunings(tunings []tunings.Tuning) []tunings.Tuning {
	tunings = h.filterUnexpectedTunings(tunings)
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

func (h *helper) filterUnexpectedTunings(list []tunings.Tuning) []tunings.Tuning {
	filtered := []tunings.Tuning{}
	for _, tuning := range list {
		if tuning.Name != "" {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}
func (h *helper) filteredByKeyword(list []tunings.Tuning) []tunings.Tuning {
	result, err := h.searchTunings(list)
	if err != nil {
		log.Errorf("tunings(%s): failed to search tunings: %v", h.reqId, err)
		return list
	}

	tuningMap := genTuningMap(list)
	filtered := []tunings.Tuning{}
	for _, hit := range result.Hits {
		filtered = append(filtered, tuningMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchTunings(tunings []tunings.Tuning) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("tunings(%s): failed to create tuning searcher: %v", h.reqId, err)
		return nil, err
	}

	for _, tuning := range tunings {
		err := searcher.Index(tuning.IndexKey(), tuning.GenSearchableOject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	keyword := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(keyword))
}

func (h *helper) filteredByHosts(list []tunings.Tuning) []tunings.Tuning {
	filtered := []tunings.Tuning{}
	for _, tuning := range list {
		if h.containsHosts(tuning.Hosts) {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}

func (h *helper) filteredByModified(list []tunings.Tuning) []tunings.Tuning {
	filtered := []tunings.Tuning{}
	for _, tuning := range list {
		if slices.Contains(h.modified, tuning.IsModified) {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}

func (h *helper) containsHosts(hosts []nodes.Host) bool {
	hostSet := make(map[string]struct{}, len(hosts))
	for _, h := range hosts {
		hostSet[h.Name] = struct{}{}
	}

	for _, h := range h.hosts {
		_, found := hostSet[h]
		if found {
			return true
		}
	}

	return false
}

func genTuningMap(list []tunings.Tuning) map[string]tunings.Tuning {
	tuningMap := map[string]tunings.Tuning{}
	for _, tuning := range list {
		tuningMap[tuning.IndexKey()] = tuning
	}

	return tuningMap
}

func (h *helper) enrichTunings(tunings *[]tunings.Tuning) {
	h.syncUpdates(tunings)
	h.sortTunings(tunings)
}

func (h *helper) syncUpdates(tunings *[]tunings.Tuning) {
	for i, tuning := range *tunings {
		(*tunings)[i] = h.syncUpdateStatus(tuning)
	}
}

func (h *helper) syncUpdateStatus(tuning tunings.Tuning) tunings.Tuning {
	tuning.SetOk()
	if !h.hasUpdateHistory(tuning) {
		return tuning
	}

	record, err := h.getUpdateRecord(tuning)
	if err != nil {
		return tuning
	}

	tuning.Status.IsUpdating = record.Status.IsUpdating
	tuning.Status.Current = record.Status.Current
	tuning.Status.UpdatedAt = record.Status.UpdatedAt
	return tuning
}

func (h *helper) hasUpdateHistory(tuning tunings.Tuning) bool {
	mongo := bsmongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		tunings.Module,
		tunings.ReqCollection(),
		bson.M{"id": tuning.GenerateId()},
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (h *helper) getUpdateRecord(tuning tunings.Tuning) (tunings.Tuning, error) {
	mongo := bsmongo.GetGlobalHelper()
	record, err := mongo.Get(
		tunings.Module,
		tunings.ReqCollection(),
		bson.M{"id": tuning.GenerateId()},
	)
	if err != nil {
		return tuning, err
	}

	t := tunings.Tuning{}
	err = record.Decode(&t)
	if err != nil {
		return tuning, err
	}

	return t, nil
}
