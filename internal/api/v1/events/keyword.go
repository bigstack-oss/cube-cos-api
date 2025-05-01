package events

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filteredByKeyword(nonFilterEvents []event.Options) []event.Options {
	if !h.isKeywordRequired() {
		return nonFilterEvents
	}

	h.setEventSearchIndex(&nonFilterEvents)
	result, err := h.searchEvents(nonFilterEvents)
	if err != nil {
		log.Errorf("events: failed to search events: %s", err.Error())
		return nonFilterEvents
	}

	eventMap := genEventMap(nonFilterEvents)
	filtered := []event.Options{}
	for _, hit := range result.Hits {
		filtered = append(filtered, eventMap[hit.ID])
	}

	return filtered
}

func (h *helper) setEventSearchIndex(nonFilterEvents *[]event.Options) {
	for i := range *nonFilterEvents {
		(*nonFilterEvents)[i].SetSearchIndex()
	}
}

func (h *helper) searchEvents(nonFilterEvents []event.Options) (*bleve.SearchResult, error) {
	searcher, err := search.New()
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
	return searcher.Search(search.WildcardQuery(h.keyword))
}

func genEventMap(nonFilterEvents []event.Options) map[string]event.Options {
	eventMap := map[string]event.Options{}
	for _, event := range nonFilterEvents {
		eventMap[event.SearchIndex] = event
	}

	return eventMap
}
