package nodes

import (
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterNodes(nodes []nodes.Node) []nodes.Node {
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
	return h.isProductRequired() ||
		h.isKeywordRequired() ||
		h.isRolesRequired() ||
		h.isLicenseStatusRequired()
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

func (h *helper) filteredByLicenseStatus(nodesToFilter []nodes.Node) []nodes.Node {
	filtered := []nodes.Node{}
	for _, node := range nodesToFilter {
		if slices.Contains(h.licenseStatuses, node.License.Status.Current) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (h *helper) filteredByProduct(nodesToFilter []nodes.Node) []nodes.Node {
	license.LowerProductsInPlace(h.products)
	filtered := []nodes.Node{}
	for _, node := range nodesToFilter {
		if slices.Contains(h.products, strings.ToLower(node.License.Product.Name)) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (h *helper) filteredByKeyword(nodesToFilter []nodes.Node) []nodes.Node {
	result, err := h.searchNodes(nodesToFilter)
	if err != nil {
		log.Errorf("nodes: failed to search nodes: %s", err.Error())
		return nodesToFilter
	}

	nodeMap := genNodeMap(nodesToFilter)
	filtered := []nodes.Node{}
	for _, hit := range result.Hits {
		filtered = append(filtered, nodeMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchNodes(nodes []nodes.Node) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("nodes: failed to create node searcher: %s", err.Error())
		return nil, err
	}

	for _, node := range nodes {
		err := searcher.Index(node.Hostname, node)
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	key := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(key))
}

func genNodeMap(nodesToFilter []nodes.Node) map[string]nodes.Node {
	nodeMap := map[string]nodes.Node{}
	for _, node := range nodesToFilter {
		nodeMap[node.Hostname] = node
	}

	return nodeMap
}

func (h *helper) filteredByRoles(nodesToFilter []nodes.Node) []nodes.Node {
	filtered := []nodes.Node{}
	for _, node := range nodesToFilter {
		if h.containsRoles(node) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (h *helper) containsRoles(node nodes.Node) bool {
	return slices.Contains(h.roles, node.Role)
}
