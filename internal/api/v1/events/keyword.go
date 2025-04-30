package events

import (
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filteredByKeyword(nonFilterEvents []events.Options) []events.Options {
	if !h.isKeywordRequired() {
		return nonFilterEvents
	}

	result, err := h.searchEvents(nonFilterEvents)
	if err != nil {
		log.Errorf("events: failed to search events: %s", err.Error())
		return nonFilterEvents
	}

	eventMap := genEventMap(nonFilterEvents)
	filtered := []events.Options{}
	for _, hit := range result.Hits {
		filtered = append(filtered, eventMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchEvents(nonFilterEvents []events.Options) (*bleve.SearchResult, error) {
	searcher, err := events.NewSearchIndex()
	if err != nil {
		log.Errorf("events: failed to create search index: %s", err.Error())
		return nil, err
	}

	for _, event := range nonFilterEvents {
		err := searcher.Index(event.SearchIndex, event.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
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
	return "*" + strings.ToLower(h.keyword) + "*"
}

func genEventMap(nonFilterEvents []events.Options) map[string]events.Options {
	eventMap := map[string]events.Options{}
	for _, event := range nonFilterEvents {
		eventMap[event.SearchIndex] = event
	}

	return eventMap
}
