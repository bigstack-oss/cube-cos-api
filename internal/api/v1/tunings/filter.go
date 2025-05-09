package tunings

import (
	"slices"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	maxSearchResults = 10000
)

func selectRolesUsingActivityAndLabels(tuningSpec *v1.TuningSpec) []*nodes.Role {
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

func getNodesBySelector(nodesToFilter []nodes.Node, selector v1.Selector) []nodes.Node {
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

func (h *helper) filterTunings(tunings []v1.Tuning) []v1.Tuning {
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

func (h *helper) filterUnexpectedTunings(tunings []v1.Tuning) []v1.Tuning {
	filtered := []v1.Tuning{}
	for _, tuning := range tunings {
		if tuning.Name != "" {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}
func (h *helper) filteredByKeyword(tunings []v1.Tuning) []v1.Tuning {
	result, err := h.searchTunings(tunings)
	if err != nil {
		log.Errorf("failed to search tunings: %s", err.Error())
		return tunings
	}

	tuningMap := genTuningMap(tunings)
	filtered := []v1.Tuning{}
	for _, hit := range result.Hits {
		filtered = append(filtered, tuningMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchTunings(tunings []v1.Tuning) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("tunings: failed to create tuning searcher: %s", err.Error())
		return nil, err
	}

	for _, tuning := range tunings {
		err := searcher.Index(tuning.SearchKey(), tuning)
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	return searcher.Search(search.WildcardQuery(h.keyword))
}

func (h *helper) filteredByHosts(tunings []v1.Tuning) []v1.Tuning {
	filtered := []v1.Tuning{}
	for _, tuning := range tunings {
		if h.containsHosts(tuning.Hosts) {
			filtered = append(filtered, tuning)
		}
	}

	return filtered
}

func (h *helper) filteredByModified(tunings []v1.Tuning) []v1.Tuning {
	filtered := []v1.Tuning{}
	for _, tuning := range tunings {
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

func genTuningMap(tunings []v1.Tuning) map[string]v1.Tuning {
	tuningMap := map[string]v1.Tuning{}
	for _, tuning := range tunings {
		tuningMap[tuning.SearchKey()] = tuning
	}

	return tuningMap
}

func (h *helper) enrichTuningPayload(tunings *[]v1.Tuning) {
	h.syncUpdates(tunings)
	h.sortTunings(tunings)
}

func (h *helper) syncUpdates(tunings *[]v1.Tuning) {
	for i, tuning := range *tunings {
		(*tunings)[i] = h.syncUpdateStatus(tuning)
	}
}

func (h *helper) syncUpdateStatus(tuning v1.Tuning) v1.Tuning {
	tuning.InitOkStatus()
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

func (h *helper) hasUpdateHistory(tuning v1.Tuning) bool {
	mongo := cubeMongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		v1.TuningDB(),
		v1.TuningReqCollection(),
		bson.M{"id": tuning.GenerateId()},
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (h *helper) getUpdateRecord(tuning v1.Tuning) (v1.Tuning, error) {
	mongo := cubeMongo.GetGlobalHelper()
	pending, err := mongo.Get(
		v1.TuningDB(),
		v1.TuningReqCollection(),
		bson.M{"id": tuning.GenerateId()},
	)
	if err != nil {
		return tuning, err
	}

	record := v1.Tuning{}
	err = pending.Decode(&record)
	if err != nil {
		return tuning, err
	}

	return record, nil
}
