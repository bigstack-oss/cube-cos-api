package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	opsimages "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	log "go-micro.dev/v5/logger"
)

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) filterImages(images []images.Image) []images.Image {
	if h.isKeywordRequired() {
		images = h.filteredByKeyword(images)
	}

	return images
}

func (h *helper) filteredByKeyword(list []images.Image) []images.Image {
	result, err := h.searchImages(list)
	if err != nil {
		log.Errorf("images: failed to search image(%v)", err)
		return list
	}

	imageMap := genImageMap(list)
	filtered := []images.Image{}
	for _, hit := range result.Hits {
		filtered = append(filtered, imageMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchImages(list []images.Image) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("images: failed to create image searcher(%v)", err)
		return nil, err
	}

	for _, image := range list {
		err := searcher.Index(image.Name, image.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	key := search.NormalizeKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(key))
}

func genImageMap(list []images.Image) map[string]images.Image {
	imageMap := map[string]images.Image{}
	for _, image := range list {
		imageMap[image.Name] = image
	}

	return imageMap
}

func (h *helper) genImageListOpts() opsimages.ListOpts {
	if h.project == "" {
		return opsimages.ListOpts{
			Owner: h.project,
		}
	}

	return opsimages.ListOpts{}
}
