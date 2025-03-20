package licenses

func (h *helper) isPageRequired() bool {
	return h.c.DefaultQuery("pageNum", "") != "" || h.c.DefaultQuery("pageSize", "") != ""
}

func (h *helper) isFilterRequired() bool {
	return h.isTypeRequired() ||
		h.isProductRequired() ||
		h.isStatusRequired() ||
		h.isKeywordRequired()
}

func (h *helper) isTypeRequired() bool {
	return h.Type != ""
}

func (h *helper) isProductRequired() bool {
	return h.Product != ""
}

func (h *helper) isStatusRequired() bool {
	return h.Status != ""
}

func (h *helper) isKeywordRequired() bool {
	return h.Keyword != ""
}
