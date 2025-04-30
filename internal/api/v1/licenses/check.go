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

func (h *helper) isAttachmentFilterRequired() bool {
	return h.isAttachmentProductRequired() || h.isKeywordRequired() || h.isAttachmentRolesRequired() || h.isAttachmenStatusRequired()
}

func (h *helper) isAttachmentProductRequired() bool {
	return h.Product != ""
}

func (h *helper) isAttachmentRolesRequired() bool {
	return len(h.Roles) > 0
}

func (h *helper) isAttachmenStatusRequired() bool {
	return len(h.Statuses) > 0
}
