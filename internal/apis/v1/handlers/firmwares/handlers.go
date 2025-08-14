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
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/firmwares/md5sum",
			Func:    uploadFirmwareMd5Sum,
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

	err = h.resetTmpFirmwareArtifacts()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.saveUploadFile()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.syncFirmwareMd5()
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

func uploadFirmwareMd5Sum(c *gin.Context) {
	h, err := initHelper(c, "uploadFirmwareMd5Sum")
	if err != nil {
		log.Errorf("images(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.resetTmpFirmwareMd5()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.saveUploadFile()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware MD5 sum uploaded successfully",
		nil,
	)
}
