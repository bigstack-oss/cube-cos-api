package licenses

func (h *helper) isPageRequired() bool {
	return h.c.DefaultQuery("pageNum", "") != "" || h.c.DefaultQuery("pageSize", "") != ""
}
