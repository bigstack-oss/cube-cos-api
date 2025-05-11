package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterFiles(sets []support.FileSet) []support.FileSet {
	if !h.isFilterRequired() {
		return sets
	}

	if h.isKeywordRequired() {
		sets = h.filteredByKeyword(sets)
	}

	if h.isRoleRequired() {
		sets = h.filteredByRoles(sets)
	}

	if h.isPeriodRequired() {
		sets = h.filteredByPeriod(sets)
	}

	return sets
}

func (h *helper) filteredByKeyword(sets []support.FileSet) []support.FileSet {
	result, err := h.searchFileSets(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to search supportFiles: %s", h.reqId, err.Error())
		return sets
	}

	setMap := h.genFileSetMap(sets)
	filtered := []support.FileSet{}
	for _, hit := range result.Hits {
		filtered = append(filtered, setMap[hit.ID])
	}

	return filtered
}

func (h *helper) filteredByRoles(fileSets []support.FileSet) []support.FileSet {
	filtered := []support.FileSet{}
	for _, fileSet := range fileSets {
		if fileSet.IncludeRoles(h.roles) {
			filtered = append(filtered, fileSet)
		}
	}

	return filtered
}

func (h *helper) filteredByPeriod(fileSets []support.FileSet) []support.FileSet {
	filtered := []support.FileSet{}
	for _, fileSet := range fileSets {
		if h.Period.InBetween(fileSet.Status.CreatedAt) {
			filtered = append(filtered, fileSet)
		}
	}

	return filtered
}

func (h *helper) searchFileSets(fileSets []support.FileSet) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to create searcher: %s", h.reqId, err.Error())
		return nil, err
	}

	for _, fileSet := range fileSets {
		err := searcher.Index(fileSet.Name, fileSet.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	keyword := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(keyword))
}

func (h *helper) genFileSetMap(files []support.FileSet) map[string]support.FileSet {
	setMap := map[string]support.FileSet{}
	for _, file := range files {
		setMap[file.Name] = file
	}

	return setMap
}
