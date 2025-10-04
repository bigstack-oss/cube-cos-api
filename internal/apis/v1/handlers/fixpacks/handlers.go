package fixpacks

import (
	"fmt"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
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
			Path:    "/fixpacks/continueAnyway/:nodeName",
			Func:    continueInterruptedFixpackUpdate,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/fixpacks/updateProgress",
			Func:    getFixpackUpdateProgress,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/fixpacks/:nodeName/version",
			Func:    getLatestNodeFixpackInfo,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/fixpacks/:version/rollback",
			Func:    rollbackFixpack,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/fixpacks/:version/rollbackableNodes",
			Func:    listRollbackableNodes,
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

	err = h.checkIfHasProcessingPkg()
	if err != nil {
		bodies.SetConflict(c, err)
		return
	}

	err = h.setPkgAs(status.Uploading)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	defer h.clearPkgBy(status.Uploading)
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

	err = h.checkFixpackDuplication()
	if err != nil {
		bodies.SetConflict(c, err)
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

	err = h.checkIfHasProcessingPkg()
	if err != nil {
		bodies.SetConflict(c, err)
		return
	}

	err = h.setPkgAs(status.Uploading)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	defer h.clearPkgBy(status.Uploading)
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

	err = h.checkIfHasProcessingPkg()
	if err != nil {
		bodies.SetConflict(c, err)
		return
	}

	err = h.setPkgAs(status.Verifying)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	defer h.clearPkgBy(status.Verifying)
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

	nodes, err := h.listUpdatableNodes(h.reqOpts.Version)
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

	h.requestOperation()
	bodies.SetAccepted(
		c,
		"Fixpack installation started successfully",
	)
}

func getFixpackUpdateProgress(c *gin.Context) {
	h, err := initHelper(c, "getFixpackUpdateProgress")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	updateProgress, err := h.getFixpackUpdateProgress()
	if err == nil {
		bodies.SetOk(c, "List of fixpack update progress", updateProgress)
		return
	}

	if err.Error() == "no fixpack history found" {
		bodies.SetOk(c, "No fixpack update history found", &update{Progresses: []progress{}})
		return
	}

	bodies.SetInternalServerError(c, err)
}

func listRollbackableNodes(c *gin.Context) {
	h, err := initHelper(c, "listRollbackableNodes")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	nodes, err := h.listRollbackableNodes()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of rollbackable nodes",
		nodes,
	)
}

func getLatestNodeFixpackInfo(c *gin.Context) {
	h, err := initHelper(c, "getLatestNodeFixpackInfo")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	info, err := cubecos.GetLatestFixpackInfo()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		fmt.Sprintf("Fixpack version of node %s", h.reqOpts.Hostname),
		info,
	)
}

func rollbackFixpack(c *gin.Context) {
	h, err := initHelper(c, "rollbackFixpack")
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

	h.requestOperation()
	bodies.SetAccepted(
		c,
		"Fixpack rollback started successfully",
	)
}

func continueInterruptedFixpackUpdate(c *gin.Context) {
	h, err := initHelper(c, "continueInterruptedFixpackUpdate")
	if err != nil {
		log.Errorf("fixpacks(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkConditionForContinue()
	if err != nil {
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

	if !h.isFixpackExists() {
		bodies.SetNotFound(c, fmt.Errorf("fixpack version %s not found", h.reqOpts.Version))
		return
	}

	if !h.isFixpackRemovable() {
		bodies.SetConflict(c, fmt.Errorf("fixpack version %s is not removable", h.reqOpts.Version))
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

	nodes, err := h.listUpdatableNodes(h.reqOpts.Version)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.updateFixpackTask(nodes)
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
