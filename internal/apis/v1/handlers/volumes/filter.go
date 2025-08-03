package volumes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) filterVolumes(volumes []volumes.Volume) []volumes.Volume {
	if h.isKeywordRequired() {
		volumes = h.filteredByKeyword(volumes)
	}

	return volumes
}

func (h *helper) filteredByKeyword(list []volumes.Volume) []volumes.Volume {
	result, err := h.searchVolumes(list)
	if err != nil {
		log.Errorf("volumes(%s): failed to search volumes(%v)", h.keyword, err)
		return list
	}

	volumeMap := volumeMap(list)
	filtered := []volumes.Volume{}
	for _, hit := range result.Hits {
		filtered = append(filtered, volumeMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchVolumes(list []volumes.Volume) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("volumes(%s): failed to create volume searcher: %v", h.keyword, err)
		return nil, err
	}

	for _, volume := range list {
		err := searcher.Index(volume.Id, volume.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	key := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(key))
}

func volumeMap(list []volumes.Volume) map[string]volumes.Volume {
	volumeMap := map[string]volumes.Volume{}
	for _, volume := range list {
		volumeMap[volume.Id] = volume
	}

	return volumeMap
}
