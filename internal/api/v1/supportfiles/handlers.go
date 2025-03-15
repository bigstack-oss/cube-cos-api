package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var Handlers = []api.Handler{
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
		Path:    "/supportFiles/:id",
		Func:    updateSupportFile,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/supportFiles/:id",
		Func:    getSupportFile,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/supportFiles/:id",
		Func:    deleteSupportFile,
	},
}

func listSupportFiles(c *gin.Context) {
	h, err := initReqHandler(c, "listSupportFiles")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", h.handler, err)
		return
	}

	api.SetStatusOk(
		c,
		"retrieved support files successfully",
		"",
	)
}

func createSupportFile(c *gin.Context) {
	h, err := initReqHandler(c, "createSupportFile")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", h.handler, err)
		return
	}

	api.SetStatusOk(
		c,
		"created support file successfully",
		"",
	)
}

func updateSupportFile(c *gin.Context) {
	h, err := initReqHandler(c, "updateSupportFileTask")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", h.handler, err)
		return
	}

	api.SetStatusOk(
		c,
		"updated support file task successfully",
		"",
	)
}

func getSupportFile(c *gin.Context) {
	h, err := initReqHandler(c, "getSupportFile")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", h.handler, err)
		return
	}

	api.SetStatusOk(
		c,
		"retrieved support file successfully",
		"",
	)
}

func deleteSupportFile(c *gin.Context) {
	h, err := initReqHandler(c, "deleteSupportFile")
	if err != nil {
		log.Infof("supportFiles(%s): failed to init req helper: %v", h.handler, err)
		return
	}

	api.SetStatusOk(
		c,
		"deleted support file successfully",
		"",
	)
}
