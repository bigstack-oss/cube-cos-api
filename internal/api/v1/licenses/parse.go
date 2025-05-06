package licenses

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listLicenses":
		return h.parseListLicenseParams()
	case "listLicenseAttachments":
		return h.parseListAttachmentParams()
	}

	return nil
}

func (h *helper) parseListLicenseParams() error {
	h.parseType()
	h.parseProducts()
	h.parseStatuses()
	h.parseKeyword()

	err := h.parseWatch()
	if err != nil {
		return err
	}

	err = h.parsePage()
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseListAttachmentParams() error {
	h.parseProduct()
	h.parseKeyword()
	h.parseRoles()
	h.parseStatuses()
	return nil
}

func (h *helper) parseType() {
	h.types = h.c.QueryArray("types")
}

func (h *helper) parseProduct() {
	h.product = h.c.DefaultQuery("product", "")
}

func (h *helper) parseProducts() {
	h.products = h.c.QueryArray("products")
}

func (h *helper) parseRoles() {
	h.roles = h.c.QueryArray("roles")
}

func (h *helper) parseStatuses() {
	h.statuses = h.c.QueryArray("statuses")
}

func (h *helper) parsePage() error {
	var err error
	h.page, err = query.GetPage(h.c)
	return err
}

func (h *helper) parseKeyword() {
	h.keyword = h.c.DefaultQuery("keyword", "")
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	return err
}
