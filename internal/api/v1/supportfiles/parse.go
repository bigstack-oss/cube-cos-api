package supportfiles

func (h *helper) parseHosts() error {
	err := h.c.ShouldBindJSON(&h.SupportFileRequest)
	if err != nil {
		return err
	}

	h.SupportFile.SetRoleByHosts(h.SupportFileRequest.Hosts)
	return nil
}
