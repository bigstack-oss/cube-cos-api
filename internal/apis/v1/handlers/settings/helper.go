package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	mongo *bsmongo.Helper
	http  *http.Helper

	task           *settings.Setting
	hostname       string
	trial          *email.Trial
	emailSender    string
	recipientEmail string
	slackChannel   string
	rawBody        []byte
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
		mongo:   bsmongo.GetGlobalHelper(),
		http:    http.GetGlobalHelper(),
		rawBody: bodies.ParseReq(c),
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listSettings() (*settings.Api, error) {
	cosSchema, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Infof("settings(%s): failed to get settings: %v", h.reqId, err)
		return nil, err
	}

	setting := cosSchema.ToApiSchema()
	h.syncUpdatingStatus(&setting)
	h.hideSenderPassword(&setting.Email.Senders)
	h.syncSenderVerification(&setting.Email.Senders)
	return &setting, nil
}

func (h *helper) updateToControllers() {
	h.updateLocal()
	if cubecos.IsVirtualIpOwner(base.Hostname) {
		h.updatePeerControllers()
	}
}
