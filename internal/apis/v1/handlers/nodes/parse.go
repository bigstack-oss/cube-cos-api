package nodes

import (
	"fmt"
	"strings"

	query "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listNodes":
		return h.parseListOptions()
	case "getNode":
		return h.parseGetOptions()
	default:
		return fmt.Errorf(
			"unknown node handler: %s",
			h.handler,
		)
	}
}

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
	h.page, err = query.GetPage(h.c)
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
