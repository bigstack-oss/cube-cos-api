package tunings

import (
	"fmt"
	"strconv"
	"strings"
)

func (h *helper) parsePage() error {
	if !h.isPageReceived() {
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

func (h *helper) parseScope() {
	h.allNodes = h.c.DefaultQuery("allNodes", "false") == "true"
}

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseHosts() {
	h.hosts = h.c.QueryArray("host")
}

func (h *helper) parseModified() {
	h.modified = h.c.DefaultQuery("modified", "false") == "true"
}

func (h *helper) parseWatch() {
	h.watch = h.c.DefaultQuery("watch", "false") == "true"
}

func (h *helper) isPageReceived() bool {
	return h.c.DefaultQuery("pageNum", "") != "" || h.c.DefaultQuery("pageSize", "") != ""
}

func (h *helper) isFilterRequired() bool {
	return h.isKeywordRequired() || h.isHostsRequired() || h.isModifiedRequired()
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isHostsRequired() bool {
	return len(h.hosts) > 0
}

func (h *helper) isModifiedRequired() bool {
	_, required := h.c.GetQuery("modified")
	return required
}
