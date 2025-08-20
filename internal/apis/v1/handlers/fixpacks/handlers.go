package fixpacks

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/fixpacks",
			Func:    listFixpacks,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/fixpacks",
			Func:    uploadFixpack,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/fixpacks/md5sum",
			Func:    uploadFixpackMd5Sum,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/fixpacks/md5sum/verify",
			Func:    verfiyFixpackAndMd5Sum,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/fixpacks/:version",
			Func:    deleteFixpack,
		},
	}
)

func listFixpacks(c *gin.Context) {
	h, err := initHelper(c, "listFixpacks")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	fixpacks, err := h.listFixpacks()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of fixpacks",
		fixpacks,
	)
}

func uploadFixpack(c *gin.Context) {
	h, err := initHelper(c, "uploadFixpack")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.resetTmpFixpackArtifacts()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.saveUploadFile()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.syncFixpackMd5()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Fixpack uploaded successfully",
		nil,
	)
}

func uploadFixpackMd5Sum(c *gin.Context) {
	h, err := initHelper(c, "uploadFixpackMd5Sum")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.resetTmpFixpackMd5()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.saveUploadFile()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	bodies.SetOk(
		c,
		"Fixpack MD5 sum uploaded successfully",
		nil,
	)
}

func verfiyFixpackAndMd5Sum(c *gin.Context) {
	h, err := initHelper(c, "verfiyFixpackAndMd5Sum")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	result, err := h.verifyFixpackAndMd5()
	if err != nil {
		bodies.SetBadRequest(c, err, result)
		return
	}

	err = h.setValidFixpack()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Fixpack and MD5 sum verified successfully",
		result,
	)
}

func deleteFixpack(c *gin.Context) {
	h, err := initHelper(c, "deleteFixpack")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.deleteFixpack()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Fixpack deleted successfully",
		nil,
	)
}
