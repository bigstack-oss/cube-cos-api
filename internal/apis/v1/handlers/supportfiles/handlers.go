package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
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
			Method:  "PATCH",
			Path:    "/supportFiles/:supportFileGroup",
			Func:    updateSupportFileTask,
		},
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/supportFiles/hosts/:hostname",
			Func:    listHostSupportFiles,
		},
	}
)

func listSupportFiles(c *gin.Context) {
	h, err := initHepler(c, "listSupportFiles")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper: %v", queries.GetReqId(c), err)
		return
	}

	supportFiles, err := h.listSupportFiles()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list support files: %v", queries.GetReqId(c), err)
		return
	}

	bodies.SetOk(
		c,
		"retrieved support files successfully",
		supportFiles,
	)
}

func createSupportFile(c *gin.Context) {
	h, err := initHepler(c, "createSupportFile")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", queries.GetReqId(c), err)
		return
	}

	h.delegateSupportFileReq()
	bodies.SetAccepted(
		c,
		"support file creation request received",
	)
}

func downloadSupportFile(c *gin.Context) {
	h, err := initHepler(c, "downloadSupportFile")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper: %v", queries.GetReqId(c), err)
		return
	}

	err = h.downloadSupportFile()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to download support file: %v", queries.GetReqId(c), err)
		return
	}
}

func updateSupportFileTask(c *gin.Context) {
	h, err := initHepler(c, "updateSupportFileTask")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper: %v", queries.GetReqId(c), err)
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
		log.Errorf("supportFiles(%s): failed to init req helper: %v", queries.GetReqId(c), err)
		return
	}

	hostSupportFiles, err := h.listHostSupportFiles()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list host support file: %v", queries.GetReqId(c), err)
		return
	}

	bodies.SetOk(
		c,
		"retrieved host support files successfully",
		hostSupportFiles,
	)

}
