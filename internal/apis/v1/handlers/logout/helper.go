package logout

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c *gin.Context
}

func initHelper(c *gin.Context) *helper {
	return &helper{c: c}
}

func (h *helper) getSession() (*samlsp.Session, error) {
	session, err := saml.SpAuth.Session.GetSession(h.c.Request)
	if err != nil {
		log.Errorf("logout(%s): failed to get session for logout: %s", queries.GetReqId(h.c), err.Error())
		return nil, err
	}

	return &session, nil
}
