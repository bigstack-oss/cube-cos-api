package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/supportfiles"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = supportfiles.ReqQueue
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/supportFiles",
			Func:    listSupportFiles,
		},
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/supportFiles/hosts/:hostname",
			Func:    listHostSupportFiles,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/supportFiles",
			Func:    createSupportFile,
		},
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/supportFiles/:supportFileGroup/:supportFileName",
			Func:    downloadSupportFile,
		},
		{
			Version: apis.V1,
			Method:  "DELETE",
			Path:    "/supportFiles/:supportFileGroup",
			Func:    deleteSupportFileGroup,
		},
		{
			Version: apis.V1,
			Method:  "DELETE",
			Path:    "/supportFiles/:supportFileGroup/:supportFileName",
			Func:    deleteSupportFile,
		},
		{
			Version: apis.V1,
			Method:  "PATCH",
			Path:    "/supportFiles/:supportFileGroup",
			Func:    updateSupportFileTask,
		},
	}
)

func listSupportFiles(c *gin.Context) {
	h, err := initHepler(c, "listSupportFiles")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init list helper(%v)", h.reqId, err)
		return
	}

	files, err := h.listSupportFiles()
	if err != nil {
		return
	}

	bodies.SetOk(
		c,
		"retrieved support files successfully",
		files,
	)
}

func createSupportFile(c *gin.Context) {
	h, err := initHepler(c, "createSupportFile")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper(%v)", h.reqId, err)
		return
	}

	err = h.checkHostValidation()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to delegate support file request(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.setSupportFileReq()
	h.delegateSupportFileReq()
	bodies.SetAccepted(
		c,
		"support file creation request received",
	)
}

func downloadSupportFile(c *gin.Context) {
	h, err := initHepler(c, "downloadSupportFile")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper(%v)", h.reqId, err)
		return
	}

	err = h.downloadSupportFile()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to download support file(%v)", h.reqId, err)
		return
	}
}

func deleteSupportFileGroup(c *gin.Context) {
	h, err := initHepler(c, "deleteSupportFileGroup")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper(%v)", h.reqId, err)
		return
	}

	err = h.deleteSupportFileGroup()
	if err != nil {
		return
	}

	bodies.SetOk(
		c,
		"support file group deleted successfully",
		nil,
	)
}

func deleteSupportFile(c *gin.Context) {
	h, err := initHepler(c, "deleteSupportFile")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper(%v)", h.reqId, err)
		return
	}

	err = cubecos.DeleteSupportFile(h.file)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to delete support file(%v)", h.reqId, err)
		return
	}

	bodies.SetOk(
		c,
		"support file deleted successfully",
		nil,
	)
}

func updateSupportFileTask(c *gin.Context) {
	h, err := initHepler(c, "updateSupportFileTask")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper(%v)", h.reqId, err)
		return
	}

	h.updateSupportFileTask()
	bodies.SetAccepted(
		c,
		"support file task updated successfully",
	)
}

func listHostSupportFiles(c *gin.Context) {
	h, err := initHepler(c, "listHostSupportFiles")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper(%v)", h.reqId, err)
		return
	}

	files, err := cubecos.ListHostSupportFiles(support.ListFileOptions{Host: h.host})
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list host support file(%v)", h.reqId, err)
		return
	}

	bodies.SetOk(
		c,
		"retrieved host support files successfully",
		files,
	)

}
