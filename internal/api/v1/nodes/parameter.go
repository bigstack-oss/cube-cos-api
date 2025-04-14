package nodes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
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
	num := h.c.DefaultQuery("pageNum", "")
	size := h.c.DefaultQuery("pageSize", "")
	if !isPageReceived(num, size) {
		return nil
	}

	if num == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if size == "" {
		return fmt.Errorf("pageSize should be provided if pageNum is provided")
	}

	var err error
	h.Page.Number, err = strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", num)
	}

	h.Page.Size, err = strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", size)
	}

	if h.Page.Number <= 0 {
		return fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if h.Page.Size <= 0 {
		return fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return nil
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = api.ParseWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}
