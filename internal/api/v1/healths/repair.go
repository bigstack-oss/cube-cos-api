package healths

func (h *helper) requestForceRepair() {
	req := genForceRepairReq(*h.module)
	reqQueue.Add(req)
}

func (h *helper) requestCheckRepair() {
	req := genCheckRepairReq()
	reqQueue.Add(req)
}
