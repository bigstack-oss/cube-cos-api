package licenses

import (
	"slices"
	"strings"

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

func (h *helper) filterLicenseAttachments(attachments []v1.LicenseAttachment) []v1.LicenseAttachment {
	if !h.isAttachmentFilterRequired() {
		return attachments
	}

	if h.isKeywordRequired() {
		attachments = h.filteredAttachmentByKeyword(attachments)
	}

	if h.isAttachmentRolesRequired() {
		attachments = h.filteredAttachmentsByRoles(attachments)
	}

	if h.isAttachmenStatusRequired() {
		attachments = h.filteredAttachmentsByStatuses(attachments)
	}

	return attachments
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
		if slices.Contains(h.Statuses, license.Status.Current) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByType(licenses []v1.License) []v1.License {
	filtered := []v1.License{}
	for _, license := range licenses {
		if slices.Contains(h.Types, license.Type) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) filteredByProduct(licenses []v1.License) []v1.License {
	v1.ToLowerInPlace(h.Products)
	filtered := []v1.License{}
	for _, license := range licenses {
		if slices.Contains(h.Products, strings.ToLower(license.Product.Name)) {
			filtered = append(filtered, license)
		}
	}

	return filtered
}

func (h *helper) searchLicenses(licenses []v1.License) (*bleve.SearchResult, error) {
	searcher := v1.GetLicenseSearcher()
	for _, license := range licenses {
		err := searcher.Index(license.Issue.Date, license.GenSearchableObject())
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
	return "*" + strings.ToLower(h.Keyword) + "*"
}

func genLicenseMap(licenses []v1.License) map[string]v1.License {
	licenseMap := map[string]v1.License{}
	for _, license := range licenses {
		licenseMap[license.Issue.Date] = license
	}

	return licenseMap
}

func (h *helper) filteredAttachmentByKeyword(attachments []v1.LicenseAttachment) []v1.LicenseAttachment {
	result, err := h.searchLicenseAttachments(attachments)
	if err != nil {
		log.Errorf("failed to search license attachments: %s", err.Error())
		return attachments
	}

	attachmentMap := genAttachmentMap(attachments)
	filtered := []v1.LicenseAttachment{}
	for _, hit := range result.Hits {
		filtered = append(filtered, attachmentMap[hit.ID])
	}

	return filtered
}

func (h *helper) searchLicenseAttachments(attachments []v1.LicenseAttachment) (*bleve.SearchResult, error) {
	searcher, err := v1.NewLicenseSearcher()
	if err != nil {
		log.Errorf("licenses: failed to create license searcher: %s", err.Error())
		return nil, err
	}

	for _, attachment := range attachments {
		err := searcher.Index(attachment.Hostname, attachment.GenSearchableObject())
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

func genAttachmentMap(attachments []v1.LicenseAttachment) map[string]v1.LicenseAttachment {
	attachmentMap := map[string]v1.LicenseAttachment{}
	for _, attachment := range attachments {
		attachmentMap[attachment.Hostname] = attachment
	}

	return attachmentMap
}

func (h *helper) filteredAttachmentsByRoles(attachments []v1.LicenseAttachment) []v1.LicenseAttachment {
	filtered := []v1.LicenseAttachment{}
	for _, attachment := range attachments {
		if slices.Contains(h.Roles, attachment.Role) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}

func (h *helper) filteredAttachmentsByStatuses(attachments []v1.LicenseAttachment) []v1.LicenseAttachment {
	filtered := []v1.LicenseAttachment{}
	for _, attachment := range attachments {
		if slices.Contains(h.Statuses, attachment.Status) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}
