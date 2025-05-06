package licenses

func (h *helper) isFilterRequired() bool {
	return h.isTypeRequired() ||
		h.isProductRequired() ||
		h.isStatusRequired() ||
		h.isKeywordRequired()
}

func (h *helper) isTypeRequired() bool {
	return len(h.types) > 0
}

func (h *helper) isProductRequired() bool {
	return len(h.products) > 0
}

func (h *helper) isStatusRequired() bool {
	return len(h.statuses) > 0
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isAttachmentFilterRequired() bool {
	return h.isAttachmentProductRequired() || h.isKeywordRequired() || h.isAttachmentRolesRequired() || h.isAttachmenStatusRequired()
}

func (h *helper) isAttachmentProductRequired() bool {
	return h.product != ""
}

func (h *helper) isAttachmentRolesRequired() bool {
	return len(h.roles) > 0
}

func (h *helper) isAttachmenStatusRequired() bool {
	return len(h.statuses) > 0
}
