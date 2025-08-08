package supportfiles

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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
		log.Errorf("supportFiles(%s): failed to search supportFiles: %v", h.reqId, err)
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
		log.Errorf("supportFiles(%s): failed to create searcher: %v", h.reqId, err)
		return nil, err
	}

	for _, fileSet := range fileSets {
		err := searcher.Index(fileSet.Name, fileSet.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	keyword := search.NormalizeKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(keyword))
}

func (h *helper) genFileSetMap(files []support.FileSet) map[string]support.FileSet {
	setMap := map[string]support.FileSet{}
	for _, file := range files {
		setMap[file.Name] = file
	}

	return setMap
}

func (h *helper) checkHostValidation() error {
	if len(h.fileReq.Hosts) == 0 {
		return fmt.Errorf("no hosts specified for support file request")
	}

	for _, host := range h.fileReq.Hosts {
		node, err := nodes.Get(host)
		if err != nil {
			return fmt.Errorf(
				"failed to get node info for %s(%v)",
				host, err,
			)
		}

		if !node.IsUp() {
			return fmt.Errorf(
				"specified hosts %s is not up, the operation is not allowed",
				node.Hostname,
			)
		}
	}

	return nil
}
