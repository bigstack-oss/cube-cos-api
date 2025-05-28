package events

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filteredByKeyword(nonFilterEvents []events.Event) []events.Event {
	if !h.isKeywordRequired() {
		return nonFilterEvents
	}

	h.setEventSearchIndex(&nonFilterEvents)
	result, err := h.searchEvents(nonFilterEvents)
	if err != nil {
		log.Errorf("events: failed to search events(%v)", err)
		return nonFilterEvents
	}

	eventMap := genEventMap(nonFilterEvents)
	filtered := []events.Event{}
	for _, hit := range result.Hits {
		filtered = append(filtered, eventMap[hit.ID])
	}

	return filtered
}

func (h *helper) setEventSearchIndex(nonFilterEvents *[]events.Event) {
	for i := range *nonFilterEvents {
		(*nonFilterEvents)[i].SetSearchIndex()
	}
}

func (h *helper) searchEvents(nonFilterEvents []events.Event) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("events: failed to create search index(%v)", err)
		return nil, err
	}

	for _, event := range nonFilterEvents {
		err := searcher.Index(event.SearchIndex, event.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	keyword := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(keyword))
}

func genEventMap(nonFilterEvents []events.Event) map[string]events.Event {
	eventMap := map[string]events.Event{}
	for _, event := range nonFilterEvents {
		eventMap[event.SearchIndex] = event
	}

	return eventMap
}
