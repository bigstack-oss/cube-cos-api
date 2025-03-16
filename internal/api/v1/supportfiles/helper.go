package supportfiles

import (
	"errors"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	handler string

	keyword string
	v1.SupportFile
	v1.SupportFileRequest
	v1.Page
	v1.Period

	watch bool
}

func initReqHandler(c *gin.Context, handler string) (*helper, error) {
	h := helper{c: c, handler: handler}
	switch h.handler {
	case "listSupportFiles":
		return initListHelper(&h)
	case "createSupportFile":
		return initCreateHelper(&h)
	case "getSupportFile":
		return initGetHelper(&h)
	case "updateSupportFile":
		return initUpdateHelper(&h)
	case "deleteSupportFile":
		return initDeleteHelper(&h)
	}

	return nil, errors.New("handler not found")
}

func initListHelper(h *helper) (*helper, error) {
	return h, nil
}

func initCreateHelper(h *helper) (*helper, error) {
	return h, h.parseHosts()
}

func initGetHelper(h *helper) (*helper, error) {
	return h, nil
}

func initUpdateHelper(h *helper) (*helper, error) {
	return h, nil
}

func initDeleteHelper(h *helper) (*helper, error) {
	return h, nil
}
