package nodes

import (
	"slices"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

const (
	maxSearchResults = 10000
)

func (h *helper) filterNodes(nodes []definition.Node) []definition.Node {
	if !h.isFilterRequired() {
		return nodes
	}

	if h.isProductRequired() {
		nodes = h.filteredByProduct(nodes)
	}

	if h.isKeywordRequired() {
		nodes = h.filteredByKeyword(nodes)
	}

	if h.isRolesRequired() {
		nodes = h.filteredByRoles(nodes)
	}

	if h.isLicenseStatusRequired() {
		nodes = h.filteredByLicenseStatus(nodes)
	}

	return nodes
}

func (h *helper) isFilterRequired() bool {
	return h.isProductRequired() || h.isKeywordRequired() || h.isRolesRequired() || h.isLicenseStatusRequired()
}

func (h *helper) isRolesRequired() bool {
	return len(h.roles) > 0
}

func (h *helper) isProductRequired() bool {
	return len(h.products) > 0
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isLicenseStatusRequired() bool {
	return len(h.licenseStatuses) > 0
}

func (h *helper) filteredByLicenseStatus(nodes []definition.Node) []definition.Node {
	filtered := []definition.Node{}
	for _, node := range nodes {
		if slices.Contains(h.licenseStatuses, node.License.Status.Current) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (h *helper) filteredByProduct(nodes []definition.Node) []definition.Node {
	filtered := []definition.Node{}
	for _, node := range nodes {
		if slices.Contains(h.products, node.License.Product.Name) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (h *helper) filteredByKeyword(nodes []definition.Node) []definition.Node {
	result, err := h.searchNodes(nodes)
	if err != nil {
		log.Errorf("nodes: failed to search nodes: %s", err.Error())
		return nodes
	}

	nodeMap := genNodeMap(nodes)
	filtered := []definition.Node{}
	for _, hit := range result.Hits {
		filtered = append(filtered, nodeMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchNodes(nodes []definition.Node) (*bleve.SearchResult, error) {
	searcher := definition.GetNodeSearcher()
	for _, node := range nodes {
		err := searcher.Index(node.Hostname, node)
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

func genNodeMap(nodes []definition.Node) map[string]definition.Node {
	nodeMap := map[string]definition.Node{}
	for _, node := range nodes {
		nodeMap[node.Hostname] = node
	}

	return nodeMap
}

func (h *helper) filteredByRoles(nodes []definition.Node) []definition.Node {
	filtered := []definition.Node{}
	for _, node := range nodes {
		if h.containsRoles(node) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (h *helper) containsRoles(node definition.Node) bool {
	return slices.Contains(h.roles, node.Role)
}
