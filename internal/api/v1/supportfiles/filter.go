package supportfiles

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filterSupportFiles(supportFiles []v1.SupportFile) []v1.SupportFile {
	if !h.isFilterRequired() {
		return supportFiles
	}

	if h.isKeywordRequired() {
		supportFiles = h.filteredByKeyword(supportFiles)
	}

	return supportFiles
}

func (h *helper) filteredByKeyword(supportFiles []v1.SupportFile) []v1.SupportFile {
	result, err := h.searchSupportFiles(supportFiles)
	if err != nil {
		log.Errorf("failed to search supportFiles: %s", err.Error())
		return supportFiles
	}

	supportFileMap := genSupportFileMap(supportFiles)
	filtered := []v1.SupportFile{}
	for _, hit := range result.Hits {
		filtered = append(filtered, supportFileMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchSupportFiles(supportFiles []v1.SupportFile) (*bleve.SearchResult, error) {
	searcher := v1.GetSupportFileSearcher()
	for _, supportFile := range supportFiles {
		err := searcher.Index(supportFile.Name, supportFile)
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

func genSupportFileMap(supportFiles []v1.SupportFile) map[string]v1.SupportFile {
	supportFileMap := map[string]v1.SupportFile{}
	for _, supportFile := range supportFiles {
		supportFileMap[supportFile.Name] = supportFile
	}

	return supportFileMap
}
