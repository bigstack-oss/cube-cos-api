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
			bleve.NewMatchQuery(h.keyword),
			maxSearchResults,
			0,
			false,
		),
	)
}

func genFileSetMap(files []support.FileSet) map[string]support.FileSet {
	fileMap := map[string]support.FileSet{}
	for _, file := range files {
		fileMap[file.Name] = file
	}

	return fileMap
}
