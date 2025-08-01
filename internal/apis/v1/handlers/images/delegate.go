package images

import "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/images"

var (
	reqQueue = images.ReqQueue
)

func (h *helper) delegateImageReq() {
	h.reqOpts.SetImporting()
	reqQueue.Add(&h.reqOpts)
}
