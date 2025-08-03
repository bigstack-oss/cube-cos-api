package volumes

import "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/volumes"

var (
	imageConvertionReqQueue = volumes.ImageConvertionReqQueue
)

func (h *helper) delegateImageConvertionReq() {
	h.imageReqOpts.SetImporting()
	imageConvertionReqQueue.Add(&h.imageReqOpts)
}
