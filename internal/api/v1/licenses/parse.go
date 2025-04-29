package licenses

import (
	"fmt"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
)

func (h *helper) parseByHandler() error {
	switch h.handler {
	case "listLicenses":
		return h.parseListParams()
	}

	return nil
}

func (h *helper) parseListParams() error {
	h.parseType()
	h.parseProduct()
	h.parseStatus()
	h.parseKeyword()
	h.parseWatch()
	return h.parsePage()
}

func (h *helper) parseType() {
	h.Types = h.c.QueryArray("types")
}

func (h *helper) parseProduct() {
	h.Products = h.c.QueryArray("products")
}

func (h *helper) parseStatus() {
	h.Statuses = h.c.QueryArray("statuses")
}

func (h *helper) parsePage() error {
	if !query.IsPageRequired(h.c) {
		return nil
	}

	num := h.c.DefaultQuery("pageNum", "")
	if num == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	size := h.c.DefaultQuery("pageSize", "")
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

func (h *helper) parseKeyword() {
	h.Keyword = h.c.DefaultQuery("keyword", "")
}

func (h *helper) parseWatch() {
	h.Watch = h.c.DefaultQuery("watch", "false") == "true"
}
