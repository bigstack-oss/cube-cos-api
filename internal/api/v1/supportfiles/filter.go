package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filterSupportFiles(fileSets []support.FileSet) []support.FileSet {
	if !h.isFilterRequired() {
		return fileSets
	}

	if h.isKeywordRequired() {
		fileSets = h.filteredByKeyword(fileSets)
	}

	if h.isRoleRequired() {
		fileSets = h.filteredByRoles(fileSets)
	}

	if h.isPeriodRequired() {
		fileSets = h.filteredByPeriod(fileSets)
	}

	return fileSets
}

func (h *helper) filteredByKeyword(fileSets []support.FileSet) []support.FileSet {
	result, err := h.searchSupportFileSets(fileSets)
	if err != nil {
		log.Errorf("failed to search supportFiles: %s", err.Error())
		return fileSets
	}

	fileSetMap := genFileSetMap(fileSets)
	filtered := []support.FileSet{}
	for _, hit := range result.Hits {
		filtered = append(filtered, fileSetMap[hit.ID])
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

func (h *helper) searchSupportFileSets(files []support.FileSet) (*bleve.SearchResult, error) {
	searcher := support.GetFileSetSearcher()
	for _, file := range files {
		err := searcher.Index(file.Name, file)
		if err != nil {
			continue
		}
	}

	return searcher.Search(
		bleve.NewSearchRequestOptions(
			bleve.NewWildcardQuery(h.wrapWilcardKeyword()),
			maxSearchResults,
			0,
			false,
		),
	)
}

func (h *helper) wrapWilcardKeyword() string {
	return "*" + h.keyword + "*"
}

func genFileSetMap(files []support.FileSet) map[string]support.FileSet {
	fileMap := map[string]support.FileSet{}
	for _, file := range files {
		fileMap[file.Name] = file
	}

	return fileMap
}
