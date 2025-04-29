package licenses

func (h *helper) isFilterRequired() bool {
	return h.isTypeRequired() ||
		h.isProductRequired() ||
		h.isStatusRequired() ||
		h.isKeywordRequired()
}

func (h *helper) isTypeRequired() bool {
	return len(h.Types) > 0
}

func (h *helper) isProductRequired() bool {
	return len(h.Products) > 0
}

func (h *helper) isStatusRequired() bool {
	return len(h.Statuses) > 0
}

func (h *helper) isKeywordRequired() bool {
	return h.Keyword != ""
}
