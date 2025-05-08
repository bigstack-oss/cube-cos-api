package nodes

import (
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
)

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseProduct() {
	h.products = h.c.QueryArray("products")
}

func (h *helper) parseRoles() {
	h.roles = h.c.QueryArray("roles")
}

func (h *helper) parseLicenseStatus() {
	h.licenseStatuses = h.c.QueryArray("licenseStatuses")
}

func (h *helper) parsePage() error {
	var err error
	h.Page, err = query.GetPage(h.c)
	return err
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}
