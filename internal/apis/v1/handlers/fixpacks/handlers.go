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
			Method:  http.MethodGet,
			Path:    "/fixpacks/:version/updatableNodes",
			Func:    listUpdatableNodes,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/fixpacks",
			Func:    installFixpack,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/fixpacks/continueAnyway",
			Func:    continueInterruptedFixpackUpdate,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/fixpacks/:version",
			Func:    deleteFixpack,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/fixpacks/tasks",
			Func:    updateFixpackTask,
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

func listUpdatableNodes(c *gin.Context) {
	h, err := initHelper(c, "listUpdatableNodes")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	nodes, err := h.listUpdatableNodes(h.version)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of updatable nodes",
		nodes,
	)
}

func installFixpack(c *gin.Context) {
	h, err := initHelper(c, "installFixpack")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkEnvConditions()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	nodes, err := h.listUpdatableNodes(h.reqOpts.Version)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.installFixpack(nodes)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"Fixpack installation started successfully",
	)
}

func continueInterruptedFixpackUpdate(c *gin.Context) {
	h, err := initHelper(c, "continueInterruptedFixpackUpdate")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.continueInterruptedFixpackUpdate()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Fixpack update continued successfully",
		nil,
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

func updateFixpackTask(c *gin.Context) {
	h, err := initHelper(c, "updateFixpackTask")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateFixpackTask()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Fixpack task updated successfully",
		nil,
	)
}
