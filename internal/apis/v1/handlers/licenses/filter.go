package licenses

import (
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterLicenses(licenses []licenses.License) []licenses.License {
	if !h.isFilterRequired() {
		return licenses
	}

	if h.isKeywordRequired() {
		licenses = h.filteredByKeyword(licenses)
	}

	if h.areStatusesRequired() {
		licenses = h.filteredByStatus(licenses)
	}

	if h.isTypeRequired() {
		licenses = h.filteredByType(licenses)
	}

	if h.areProductsRequired() {
		licenses = h.filteredByProduct(licenses)
	}

	return licenses
}

func (h *helper) filterAttachments(attachments []licenses.Attachment) []licenses.Attachment {
	if !h.isAttachmentFilterRequired() {
		return attachments
	}

	if h.isKeywordRequired() {
		attachments = h.filteredAttachmentByKeyword(attachments)
	}

	if h.areRolesRequired() {
		attachments = h.filteredAttachmentsByRoles(attachments)
	}

	if h.areStatusesRequired() {
		attachments = h.filteredAttachmentsByStatuses(attachments)
	}

	return attachments
}

func (h *helper) filteredByKeyword(list []licenses.License) []licenses.License {
	result, err := h.searchLicenses(list)
	if err != nil {
		log.Errorf("failed to search licenses: %v", err)
		return list
	}

	licenseMap := genLicenseMap(list)
	filtered := []licenses.License{}
	for _, hit := range result.Hits {
		filtered = append(filtered, licenseMap[hit.ID])
	}

	return filtered
}

func (h *helper) filteredByStatus(list []licenses.License) []licenses.License {
	filtered := []licenses.License{}
	for _, license := range list {
		if slices.Contains(h.statuses, license.Status.Current) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByType(list []licenses.License) []licenses.License {
	filtered := []licenses.License{}
	for _, license := range list {
		if slices.Contains(h.types, license.Type) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByProduct(list []licenses.License) []licenses.License {
	licenses.LowerProductsInPlace(h.products)
	filtered := []licenses.License{}
	for _, license := range list {
		if slices.Contains(h.products, strings.ToLower(license.Product.Name)) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) searchLicenses(list []licenses.License) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("licenses(%s): failed to new searcher: %v", h.reqId, err)
		return nil, err
	}

	for _, license := range list {
		err := searcher.Index(license.Issue.Date, license.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	keyword := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(keyword))
}

func genLicenseMap(list []licenses.License) map[string]licenses.License {
	licenseMap := map[string]licenses.License{}
	for _, license := range list {
		licenseMap[license.Issue.Date] = license
	}

	return licenseMap
}

func (h *helper) filteredAttachmentByKeyword(attachments []licenses.Attachment) []licenses.Attachment {
	result, err := h.searchAttachments(attachments)
	if err != nil {
		log.Errorf("failed to search license attachments: %v", err)
		return attachments
	}

	attachmentMap := genAttachmentMap(attachments)
	filtered := []licenses.Attachment{}
	for _, hit := range result.Hits {
		filtered = append(filtered, attachmentMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchAttachments(attachments []licenses.Attachment) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("licenses(%s): failed to new searcher: %v", h.reqId, err)
		return nil, err
	}

	for _, attachment := range attachments {
		err := searcher.Index(attachment.Hostname, attachment.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	key := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(key))
}

func genAttachmentMap(attachments []licenses.Attachment) map[string]licenses.Attachment {
	attachmentMap := map[string]licenses.Attachment{}
	for _, attachment := range attachments {
		attachmentMap[attachment.Hostname] = attachment
	}

	return attachmentMap
}

func (h *helper) filteredAttachmentsByRoles(attachments []licenses.Attachment) []licenses.Attachment {
	filtered := []licenses.Attachment{}
	for _, attachment := range attachments {
		if slices.Contains(h.roles, attachment.Role) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}

func (h *helper) filteredAttachmentsByStatuses(attachments []licenses.Attachment) []licenses.Attachment {
	filtered := []licenses.Attachment{}
	for _, attachment := range attachments {
		if slices.Contains(h.statuses, attachment.Status) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}
