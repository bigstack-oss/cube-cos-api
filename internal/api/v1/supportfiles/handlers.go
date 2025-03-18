package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/supportfiles"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = supportfiles.ReqQueue
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  "GET",
			Path:    "/supportFiles",
			Func:    listSupportFiles,
		},
		{
			Version: api.V1,
			Method:  "POST",
			Path:    "/supportFiles",
			Func:    createSupportFile,
		},
		{
			Version: api.V1,
			Method:  "PATCH",
			Path:    "/supportFiles/tasks/:id",
			Func:    updateSupportFileTask,
		},
		{
			Version: api.V1,
			Method:  "GET",
			Path:    "/supportFiles/:hostname",
			Func:    getSupportFile,
		},
	}
)

func listSupportFiles(c *gin.Context) {
	h, err := initReqHandler(c, "listSupportFiles")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper: %v", api.GetReqId(c), err)
		return
	}

	supportFiles, err := h.listSupportFiles()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list support files: %v", api.GetReqId(c), err)
		return
	}

	api.SetStatusOk(
		c,
		"retrieved support files successfully",
		supportFiles,
	)
}

func createSupportFile(c *gin.Context) {
	h, err := initReqHandler(c, "createSupportFile")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", api.GetReqId(c), err)
		return
	}

	h.delegateSupportFileReq()
	api.SetStatusAccepted(
		c,
		"support file creation request received",
	)
}

func updateSupportFileTask(c *gin.Context) {
	h, err := initReqHandler(c, "updateSupportFileTask")
	if err != nil {
		log.Errorf("supportFiles(%s): failed to init req helper: %v", api.GetReqId(c), err)
		return
	}

	h.updateSupportFileTask()
	api.SetStatusAccepted(
		c,
		"support file task updated successfully",
	)
}

func getSupportFile(c *gin.Context) {}
