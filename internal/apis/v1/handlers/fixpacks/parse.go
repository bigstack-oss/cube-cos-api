package fixpacks

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listFixpacks":
		return h.parseListParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get page parameters (%v)", h.reqId, err)
		return err
	}

	return nil
}
