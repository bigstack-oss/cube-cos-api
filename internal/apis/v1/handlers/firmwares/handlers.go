package firmwares

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/firmwares",
			Func:    uploadFirmware,
		},
	}
)

func uploadFirmware(c *gin.Context) {
	h, err := initHelper(c, "uploadFirmware")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.resetTmpSpace()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.saveUploadFirmware()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.syncFirmwareMd5Sum()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware uploaded successfully",
		nil,
	)
}
