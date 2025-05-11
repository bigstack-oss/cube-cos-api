package licenses

import (
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/blevesearch/bleve/v2"
	log "go-micro.dev/v5/logger"
)

func (h *helper) filterLicenses(licenses []license.Options) []license.Options {
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

func (h *helper) filterAttachments(attachments []license.Attachment) []license.Attachment {
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

func (h *helper) filteredByKeyword(licenses []license.Options) []license.Options {
	result, err := h.searchLicenses(licenses)
	if err != nil {
		log.Errorf("failed to search licenses: %s", err.Error())
		return licenses
	}

	licenseMap := genLicenseMap(licenses)
	filtered := []license.Options{}
	for _, hit := range result.Hits {
		filtered = append(filtered, licenseMap[hit.ID])
	}

	return filtered
}

func (h *helper) filteredByStatus(licenses []license.Options) []license.Options {
	filtered := []license.Options{}
	for _, license := range licenses {
		if slices.Contains(h.statuses, license.Status.Current) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByType(licenses []license.Options) []license.Options {
	filtered := []license.Options{}
	for _, license := range licenses {
		if slices.Contains(h.types, license.Type) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByProduct(licenses []license.Options) []license.Options {
	license.LowerProductsInPlace(h.products)
	filtered := []license.Options{}
	for _, license := range licenses {
		if slices.Contains(h.products, strings.ToLower(license.Product.Name)) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) searchLicenses(licenses []license.Options) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("licenses(%s): failed to new searcher: %s", queries.GetReqId(h.c), err.Error())
		return nil, err
	}

	for _, license := range licenses {
		err := searcher.Index(license.Issue.Date, license.GenSearchableObject())
		if err != nil {
			continue
		}
	}

	defer searcher.Close()
	keyword := search.NormalizedKeyword(h.keyword)
	return searcher.Search(search.WildcardQuery(keyword))
}

func genLicenseMap(licenses []license.Options) map[string]license.Options {
	licenseMap := map[string]license.Options{}
	for _, license := range licenses {
		licenseMap[license.Issue.Date] = license
	}

	return licenseMap
}

func (h *helper) filteredAttachmentByKeyword(attachments []license.Attachment) []license.Attachment {
	result, err := h.searchAttachments(attachments)
	if err != nil {
		log.Errorf("failed to search license attachments: %s", err.Error())
		return attachments
	}

	attachmentMap := genAttachmentMap(attachments)
	filtered := []license.Attachment{}
	for _, hit := range result.Hits {
		filtered = append(filtered, attachmentMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchAttachments(attachments []license.Attachment) (*bleve.SearchResult, error) {
	searcher, err := search.New()
	if err != nil {
		log.Errorf("licenses(%s): failed to new searcher: %s", queries.GetReqId(h.c), err.Error())
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

func genAttachmentMap(attachments []license.Attachment) map[string]license.Attachment {
	attachmentMap := map[string]license.Attachment{}
	for _, attachment := range attachments {
		attachmentMap[attachment.Hostname] = attachment
	}

	return attachmentMap
}

func (h *helper) filteredAttachmentsByRoles(attachments []license.Attachment) []license.Attachment {
	filtered := []license.Attachment{}
	for _, attachment := range attachments {
		if slices.Contains(h.roles, attachment.Role) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}

func (h *helper) filteredAttachmentsByStatuses(attachments []license.Attachment) []license.Attachment {
	filtered := []license.Attachment{}
	for _, attachment := range attachments {
		if slices.Contains(h.statuses, attachment.Status) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}
