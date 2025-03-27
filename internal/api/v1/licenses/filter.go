package licenses

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filterLicenses(licenses []v1.License) []v1.License {
	if !h.isFilterRequired() {
		return licenses
	}

	if h.isKeywordRequired() {
		licenses = h.filteredByKeyword(licenses)
	}

	if h.isStatusRequired() {
		licenses = h.filteredByStatus(licenses)
	}

	if h.isTypeRequired() {
		licenses = h.filteredByType(licenses)
	}

	if h.isProductRequired() {
		licenses = h.filteredByProduct(licenses)
	}

	return licenses
}

func (h *helper) filteredByKeyword(licenses []v1.License) []v1.License {
	result, err := h.searchLicenses(licenses)
	if err != nil {
		log.Errorf("failed to search licenses: %s", err.Error())
		return licenses
	}

	licenseMap := genLicenseMap(licenses)
	filtered := []v1.License{}
	for _, hit := range result.Hits {
		filtered = append(filtered, licenseMap[hit.ID])
	}

	return filtered
}

func (h *helper) filteredByStatus(licenses []v1.License) []v1.License {
	filtered := []v1.License{}
	for _, license := range licenses {
		if license.Status.Current == h.Status {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByType(licenses []v1.License) []v1.License {
	filtered := []v1.License{}
	for _, license := range licenses {
		if license.Type == h.Type {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByProduct(licenses []v1.License) []v1.License {
	filtered := []v1.License{}
	for _, license := range licenses {
		if license.Product.Name == h.Product {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) searchLicenses(licenses []v1.License) (*bleve.SearchResult, error) {
	searcher := v1.GetLicenseSearcher()
	for _, license := range licenses {
		err := searcher.Index(license.Name, license)
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
	return "*" + h.Keyword + "*"
}

func genLicenseMap(licenses []v1.License) map[string]v1.License {
	licenseMap := map[string]v1.License{}
	for _, license := range licenses {
		licenseMap[license.Name] = license
	}

	return licenseMap
}
