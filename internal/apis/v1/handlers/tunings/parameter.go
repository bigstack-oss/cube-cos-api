package tunings

import (
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
)

func (h *helper) parsePage() error {
	var err error
	h.Page, err = queries.GetPage(h.c)
	return err
}

func (h *helper) parseScope() {
	h.allNodes = h.c.DefaultQuery("allNodes", "true") == "true"
}

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseHosts() {
	h.hosts = h.c.QueryArray("host")
}

func (h *helper) parseModified() {
	modifies := h.c.QueryArray("modified")
	for _, m := range modifies {
		h.modified = append(
			h.modified,
			strings.ToLower(m) == "true",
		)
	}
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = queries.GetWatch(h.c)
	return err
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
