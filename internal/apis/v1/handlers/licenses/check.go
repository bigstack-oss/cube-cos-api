package licenses

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"

func (h *helper) isFilterRequired() bool {
	return h.isTypeRequired() ||
		h.areProductsRequired() ||
		h.areStatusesRequired() ||
		h.isKeywordRequired()
}

func (h *helper) isTypeRequired() bool {
	return len(h.types) > 0
}

func (h *helper) areProductsRequired() bool {
	return len(h.products) > 0
}

func (h *helper) areStatusesRequired() bool {
	return len(h.statuses) > 0
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isAttachmentFilterRequired() bool {
	return h.isProductRequired() ||
		h.isKeywordRequired() ||
		h.areRolesRequired() ||
		h.areStatusesRequired()
}

func (h *helper) isProductRequired() bool {
	return h.product != ""
}

func (h *helper) areRolesRequired() bool {
	return len(h.roles) > 0
}

func (h *helper) isLicenseNotInstalled(list []licenses.License) bool {
	return len(list) == 0
}
